package nuchal

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nchl/pkg/account"
	"nchl/pkg/product"
	"os"
	"time"
)

// Config for the environment
type Config struct {
	*account.Group
	*time.Duration
	*product.Strategy
	TestMode bool
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

// NewConfig reads configuration from environment variables and validates it
func NewConfig() (*Config, error) {

	log.Info().Msg("creating configuration")

	c := new(Config)

	// get a new account group
	if group, err := account.NewGroup(); err != nil {
		log.Error().Err(err)
		return nil, err
	} else {
		c.Group = group
	}

	// get a new product strategy
	if strategy, err := product.NewStrategy(); err != nil {
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

	c.TestMode = os.Getenv("MODE") == "test"
	if c.TestMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Msgf("created configuration [%v]", c)
	return c, nil
}
