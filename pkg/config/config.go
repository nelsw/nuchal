package config

import (
	"fmt"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

// Session of the application - only active while the executable is run within the defined period range.
type Session struct {

	// Port is the port where nuchal will serve (simulation) html files.
	Port int `envconfig:"PORT"`

	// Period is a range of time representing when to start and stop executing the trade command.
	Period `yaml:"period"`

	*cbp.Api `yaml:"cbp"`

	Products map[string]cbp.Product
}

func (s *Session) SimulationStart() int64 {
	start, _ := time.ParseDuration(s.Alpha)
	return start.Nanoseconds()
}

func (s *Session) SimulationAddress() string {
	return fmt.Sprintf("localhost:%d", s.Port)
}

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if os.Getenv("DEBUG") == "true" {
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

func (s *Session) User() string {
	return os.Getenv("USER")
}

// NewSession reads configuration from environment variables and validates it
func NewSession() (*Session, error) {

	util.PrintlnBanner()

	c := new(Session)

	log.Info().Msg(util.Fish + " ... hello " + c.User())
	log.Info().Msg(util.Fish + " ... let's get configured")
	log.Debug().Msg("configure session")

	err := util.ConfigFromYml(c)
	if err != nil {
		// either the file didn't exist or wasn't properly formatted
		err = util.ConfigFromEnv(c)
		if err != nil {
			return nil, err
		}
	}

	if api, err := cbp.NewApi(); err != nil {
		return nil, err
	} else {
		c.Api = api
	}

	// While trade and report commands do not requiring the database,
	// simulations do, which should occur before trading & reporting.
	if err := db.InitDb(); err != nil {
		return nil, err
	}

	// Initiating products will fetch the most recent list of
	// cryptocurrencies and apply patterns to each product.

	all, err := c.Api.GetUsdProducts()
	if err != nil {
		return nil, err
	}
	m := map[string]cb.Product{}
	for _, a := range *all {
		m[a.ID] = a
	}

	var wrapper struct {
		Patterns []cbp.Pattern `yaml:"patterns"`
	}

	if err := util.ConfigFromYml(&wrapper); err != nil {
		return nil, err
	}

	if len(wrapper.Patterns) < 1 {
		for _, product := range *all {
			pattern := new(cbp.Pattern)
			pattern.Id = product.ID
			wrapper.Patterns = append(wrapper.Patterns, *pattern)
		}
	}

	mm := map[string]cbp.Product{}
	for _, pattern := range wrapper.Patterns {
		product := m[pattern.Id]
		if pattern.Size == 0 {
			pattern.Size = util.Float64(product.BaseMinSize) * 10
		}
		if pattern.Gain == 0 {
			pattern.Gain = c.Taker * 6
		}
		if pattern.Loss == 0 {
			pattern.Loss = c.Taker * 9
		}
		if pattern.Delta == 0 {
			pattern.Delta = util.Float64(product.QuoteIncrement) * 10
		}
		mm[pattern.Id] = cbp.Product{product, pattern}
	}
	c.Products = mm

	log.Debug().Interface("session", c).Msg("configure")
	log.Info().Msg(util.Fish + " ... OKAY, we're all set")

	return c, nil
}

func (s *Session) GetTradingPositions() (*[]cbp.Position, error) {
	positions, err := s.GetActivePositions()
	if err != nil {
		return nil, err
	}

	var result []cbp.Position
	for _, position := range *positions {
		if position.Currency == "USD-USD" || position.Balance() == position.Hold() {
			continue
		}
		result = append(result, position)
	}

	return &result, nil
}

func (s *Session) GetActivePositions() (*[]cbp.Position, error) {

	positions, err := s.Api.GetActivePositions()
	if err != nil {
		return nil, err
	}

	var result []cbp.Position
	for _, position := range *positions {
		if position.Currency != "USD" {
			position.Pattern = s.Products[position.ProductId()].Pattern
		}
		result = append(result, position)
	}

	return &result, nil
}
