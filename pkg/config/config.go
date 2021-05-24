package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/nelsw/nuchal/pkg/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

// Config for the environment
type Config struct {
	Port        string `envconfig:"PORT" default:":8080"`
	Mode        string `envconfig:"MODE" default:"DEBUG"`
	DurationStr string `envconfig:"DURATION" default:"24h"`
	*model.Group
	*model.Strategy
}

func (c Config) IsTestMode() bool {
	return c.Mode == "TEST"
}

func (c Config) IsDebugMode() bool {
	return c.Mode == "DEBUG"
}

func (c *Config) StartTime() *time.Time {
	now := time.Now()
	then := now.Add(-*c.Duration())
	return &then
}

func (c *Config) Duration() *time.Duration {
	duration, _ := time.ParseDuration(c.DurationStr)
	return &duration
}

func (c *Config) StartTimeUnixNano() int64 {
	then := c.StartTime()
	nano := then.UnixNano()
	return nano
}

func (c *Config) EndTime() *time.Time {
	now := time.Now()
	then := now.Add(*c.Duration())
	return &then
}

func (c *Config) IsTimeToExit() bool {
	now := time.Now()
	then := *c.EndTime()
	return now.After(then)
}

// NewConfig reads configuration from environment variables and validates it
func NewConfig() (*Config, error) {

	c := new(Config)
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if c.IsTestMode() || c.IsDebugMode() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("***%s****", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}

	log.Info().Msg("configuring nuchal")

	log.Info().Msg("configuring users")
	if group, err := model.NewGroup(); err != nil {
		return nil, err
	} else {
		c.Group = group
		for _, user := range c.Group.Users {
			log.Debug().Str("user", user.Name).Send()
		}
		log.Info().Msg("configured users")
	}

	log.Info().Msg("configuring products")
	if strategy, err := model.NewStrategy(); err != nil {
		return nil, err
	} else {
		c.Strategy = strategy
		for _, posture := range strategy.Postures {
			log.Debug().Str("product", posture.Id).Send()
		}
		log.Info().Msgf("configured products")
	}

	// check that the duration is valid
	if _, err := time.ParseDuration(c.DurationStr); err != nil {
		return nil, err
	}

	log.Info().Msg("configured nuchal")
	return c, nil
}
