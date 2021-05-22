package config

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal/pkg/model"
	"os"
	"strings"
	"time"
)

// Config for the environment
type Config struct {
	SimPort string `envconfig:"SIM_PORT" default:":8080"`
	*model.Group
	*time.Duration
	*model.Strategy
}

func (c Config) StartTime() *time.Time {
	start := time.Now().Add(-*c.Duration)
	return &start
}

func (c Config) StartTimeUnixNano() int64 {
	return c.StartTime().UnixNano()
}

func (c Config) EndTime() *time.Time {
	end := time.Now().Add(*c.Duration)
	return &end
}

func (c Config) IsTimeToExit() bool {
	return time.Now().After(*c.EndTime())
}

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if os.Getenv("MODE") == "test" {
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
}

// NewConfig reads configuration from environment variables and validates it
func NewConfig() (*Config, error) {

	log.Info().Msg("creating configuration")

	c := new(Config)

	// get a new report group
	if group, err := model.NewGroup(); err != nil {
		log.Error().Err(err).Msg("error ")
		return nil, err
	} else {
		c.Group = group
	}

	// get a new product strategy
	if strategy, err := model.NewStrategy(); err != nil {
		log.Error().Err(err)
		return nil, err
	} else {
		c.Strategy = strategy
	}

	// set duration from environment variable or set it to a default amount
	d := os.Getenv("DURATION")
	if d == "" {
		d = "24h"
	}

	if duration, err := time.ParseDuration(d); err != nil {
		return nil, err
	} else {
		c.Duration = &duration
	}

	log.Info().Msgf("created configuration [%v]", c)
	return c, nil
}
