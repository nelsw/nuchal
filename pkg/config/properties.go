package config

import (
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/model"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

// Properties for the environment
type Properties struct {

	// Mode is a value for defining application mode.
	Mode string `envconfig:"MODE" default:"DEBUG"`

	// SimPort is the port where nuchal will serve simulation html files.
	SimPort int `envconfig:"SIM_PORT" default:"8080"`

	// DurationStr is a time.Duration parsable value for the amount of time the command should be executed.
	DurationStr string `envconfig:"DURATION" default:"24h"`

	// AlphaStr is time.Time parsable value for defining when products are eligible for simulation and trading.
	AlphaStr string `envconfig:"ALPHA"`

	// OmegaStr is time.Time parsable value for defining when products are no longer eligible for simulation or trading.
	OmegaStr string `envconfig:"OMEGA"`

	Host string `envconfig:"POSTGRES_HOST" default:"localhost"`
	User string `envconfig:"POSTGRES_USER" default:"postgres"`
	Pass string `envconfig:"POSTGRES_PASSWORD" default:"somePassword"`
	Name string `envconfig:"POSTGRES_DB" default:"nuchal"`
	Port int    `envconfig:"POSTGRES_PORT" default:"5432"`

	Users []model.User `json:"users"`

	Products map[string]model.Product
}

func (p *Properties) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", p.Host, p.User, p.Pass, p.Name, p.Port)
}

func (p *Properties) SimAddress() string {
	return fmt.Sprintf("localhost:%d", p.SimPort)
}

func (p *Properties) IsTestMode() bool {
	return p.Mode == "TEST"
}

func (p *Properties) IsDebugMode() bool {
	return p.Mode == "DEBUG"
}

func (p *Properties) Duration() *time.Duration {
	duration, _ := time.ParseDuration(p.DurationStr)
	return &duration
}

func (p *Properties) StartTimeUnixNano() int64 {
	now := time.Now()
	then := now.Add(-*p.Duration())
	nano := then.UnixNano()
	return nano
}

func (p *Properties) IsTimeToExit() bool {
	now := time.Now()
	then := now.Add(*p.Duration())
	return now.After(then)
}

// NewProperties reads configuration from environment variables and validates it
func NewProperties() (*Properties, error) {

	c := new(Properties)
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}

	c.initLogging()

	log.Info().Msg("configuring")

	if err := c.initDatabase(); err != nil {
		return nil, err
	} else if err := c.initUsers(); err != nil {
		return nil, err
	} else if err := c.initProducts(); err != nil {
		return nil, err
	} else if _, err := time.ParseDuration(c.DurationStr); err != nil {
		return nil, err
	}

	log.Info().Msg("configured")

	return c, nil
}

func (p *Properties) initLogging() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if p.IsTestMode() || p.IsDebugMode() {
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

func (p *Properties) initDatabase() error {
	log.Info().Msg("configuring database")
	if pg, err := db.OpenDB(p.DSN()); err != nil {
		return err
	} else if sql, err := pg.DB(); err != nil {
		return err
	} else if err := sql.Ping(); err != nil {
		return err
	} else if err := sql.Close(); err != nil {
		return err
	}
	log.Info().Msg("configured database")
	return nil
}

func (p *Properties) initProducts() error {

	log.Info().Msg("configuring products")

	var productsWrapper struct {
		All []cb.Product `json:"products"`
	}

	if file, err := os.Open("pkg/config/products.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open products.json")
		return err
	} else if err := json.NewDecoder(file).Decode(&productsWrapper); err != nil {
		log.Warn().Err(err).Msg("unable to decode products.json")
		return err
	}

	var patternsWrapper struct {
		All []model.Pattern `json:"tweezer"`
	}

	if file, err := os.Open("pkg/config/patterns.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open patterns.json")
		return err
	} else if err := json.NewDecoder(file).Decode(&patternsWrapper); err != nil {
		log.Warn().Err(err).Msg("unable to decode patterns.json")
		return err
	}

	productMap := map[string]cb.Product{}
	for _, product := range productsWrapper.All {
		productMap[product.ID] = product
	}

	p.Products = map[string]model.Product{}
	for _, pattern := range patternsWrapper.All {
		if pattern.Enable {
			p.Products[pattern.Id] = model.Product{productMap[pattern.Id], pattern}
			log.Debug().Str("product", pattern.Id).Send()
		}
	}

	log.Info().Msgf("configured products")

	return nil
}

func (p *Properties) initUsers() error {

	log.Info().Msg("configuring users")

	if file, err := os.Open("pkg/config/users.json"); err == nil {
		if err := json.NewDecoder(file).Decode(&p); err != nil {
			return err
		}
	} else {
		db.NewDB(p.DSN()).Where("enable = ?", true).Find(&p.Users)
	}

	if p.Users == nil && len(p.Users) < 1 {
		usr, err := model.NewUser()
		if err != nil {
			return err
		}
		p.Users = append(p.Users, *usr)
	}

	var enabledUsers []model.User
	for _, user := range p.Users {
		if !user.Enable {
			continue
		}
		if err := user.Validate(); err != nil {
			return err
		}
		enabledUsers = append(enabledUsers, user)
		log.Debug().Str("user", user.Name).Send()
	}

	p.Users = enabledUsers

	log.Info().Msg("configured users")

	return nil
}
