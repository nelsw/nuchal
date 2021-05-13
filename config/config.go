package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

// Config for the environment
type Config struct {
	Database Database  `json:"database"`
	Users    []User    `json:"users"`
	Duration string    `json:"duration"`
	Postures []Posture `json:"-"`
}

func (c Config) User() User {
	return c.Users[0]
}

type Database struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Name string `json:"name"`
	Port int    `json:"port"`
}

func (db Database) DSN() string {
	return fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%d", db.User, db.Pass, db.Name, db.Port)
}

type Position struct {
	Id     string `json:"id"`
	Gain   string `json:"gain"`
	Loss   string `json:"loss"`
	Delta  string `json:"delta"`
	Size   string `json:"size"`
	Enable bool   `json:"enable,omitempty"`
}

type Product struct {
	Id             string `json:"id"`
	BaseCurrency   string `json:"base_currency"`
	QuoteCurrency  string `json:"quote_currency"`
	BaseMinSize    string `json:"base_min_size"`
	BaseMaxSize    string `json:"base_max_size"`
	QuoteIncrement string `json:"quote_increment"`
}

type User struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Secret     string `json:"secret"`
	Enable     bool   `json:"enable"`
}

type Posture struct {
	Product
	Position
}

func (p Posture) ProductId() string {
	return p.Product.Id
}

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// NewConfig reads configuration from environment variables and validates it
func NewConfig() (*Config, error) {

	log.Info().Msg("creating configuration")

	c := new(Config)

	if file, err := os.Open("./config/users.json"); err == nil {
		log.Info().Msg("found a users config file")
		if err := json.NewDecoder(file).Decode(&c); err != nil {
			log.Warn().Err(err).Msg("unable to decode users.json")
			return nil, err
		}
	}

	if c.Duration == "" || c.Users == nil || len(c.Users) < 1 {
		name := flag.String("name", "Connor", "The first OR full name of a Coinbase Pro user")
		pass := flag.String("pass", "abcde12345", "A Coinbase Pro API Passphrase")
		key := flag.String("key", "abcdefg1234567...", "A Coinbase Pro API Key")
		secret := flag.String("secret", "abcdefghij1234567890...", "A Coinbase Pro API Secret")
		duration := flag.String("duration", "36h30m40s", "the duration to execute the command")
		flag.Parse()

		if c.Duration == "" {
			c.Duration = *duration
		}
		if c.Users == nil || len(c.Users) < 1 {
			c.Users = append(c.Users, User{
				*name,
				*key,
				*pass,
				*secret,
				true,
			})
		}
	}

	var products struct {
		All []Product `json:"products"`
	}

	if file, err := os.Open("./config/products.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open products.json")
		return nil, err
	} else if err := json.NewDecoder(file).Decode(&products); err != nil {
		log.Warn().Err(err).Msg("unable to decode products.json")
		return nil, err
	}

	var positions struct {
		All []Position `json:"positions"`
	}

	if file, err := os.Open("./config/positions.json"); err != nil {
		log.Warn().Err(err).Msg("unable to open positions.json")
		return nil, err
	} else if err := json.NewDecoder(file).Decode(&positions); err != nil {
		log.Warn().Err(err).Msg("unable to decode positions.json")
		return nil, err
	} else {
		fmt.Println(len(positions.All))
	}

	log.Info().Interface("positions", positions)

	productMap := map[string]Product{}
	for _, product := range products.All {
		productMap[product.Id] = product
	}

	for _, position := range positions.All {
		if position.Enable {
			c.Postures = append(c.Postures, Posture{productMap[position.Id], position})
		}
	}

	log.Info().Msg("created configuration")
	return c, nil
}
