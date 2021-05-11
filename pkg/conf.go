package pkg

import (
	"encoding/json"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gol "log"
	"os"
	"strings"
	"time"
)

// Config for the environment
type Config struct {
	AllCoinbaseProducts []Product `json:"all_coinbase_products"`
	SimulationProductId string    `json:"simulation_product_id"`
	TradeProductIds     []string  `json:"trade_product_ids,omitempty"`
	Users               []User    `json:"users,omitempty"`
	DatabaseUrl         string    `json:"database_url"`
}

type Product struct {
	Id             string `json:"id"`
	BaseCurrency   string `json:"base_currency"`
	QuoteCurrency  string `json:"quote_currency"`
	BaseMinSize    string `json:"base_min_size"`
	BaseMaxSize    string `json:"base_max_size"`
	QuoteIncrement string `json:"quote_increment"`
	StopGain       string `json:"stop_gain"`
	StopLoss       string `json:"stop_loss"`
	Tweezer        string `json:"tweezer"`
	Size           string `json:"size"`
}

type User struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Secret     string `json:"secret"`
}

var db *gorm.DB

// NewDefaultConfig reads configuration from environment variables and validates it
func NewDefaultConfig() *Config {
	cfg := new(Config)

	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	} else if file, err := os.Open(".conf/config.json"); err != nil {
		panic(err)
	} else if err = json.NewDecoder(file).Decode(&cfg); err != nil {
		panic(err)
	} else if db, err = gorm.Open(postgres.Open(cfg.DatabaseUrl), &gorm.Config{
		Logger: logger.New(
			gol.New(os.Stdout, "\r\n", gol.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second,   // Slow SQL threshold
				LogLevel:                  logger.Silent, // Log level
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,         // Disable color
			},
		),
	}); err != nil {
		panic(err)
	}
	if cfg.Users == nil || len(cfg.Users) < 1 {
		user := User{
			os.Getenv("name"),
			os.Getenv("key"),
			os.Getenv("pass"),
			os.Getenv("secret"),
		}
		if user == (User{}) {
			panic("no user defined")
		}
		cfg.Users = append(cfg.Users, User{
			os.Getenv("name"),
			os.Getenv("key"),
			os.Getenv("pass"),
			os.Getenv("secret"),
		})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Info().Msg("Configuration loaded")
	return cfg
}

func (p Product) stopLoss() float64 {
	return Float64(p.StopLoss)
}

func (p Product) stopGain() float64 {
	return Float64(p.StopGain)
}

func (p Product) EntryPrice(price float64) float64 {
	price += price * p.stopGain()
	return price
}

func (p Product) LossPrice(price float64) float64 {
	price -= price * p.stopLoss()
	return price
}

func (c Config) FindUserByFirstName(name string) User {
	for _, user := range c.Users {
		if strings.Contains(user.Name, name) {
			return user
		}
	}
	panic("no user found for " + name)
}

func (c Config) SimulationProduct() Product {
	for _, product := range c.AllCoinbaseProducts {
		if product.Id == c.SimulationProductId {

			return product
		}
	}
	panic("simulation product not found")
}

func Size(price float64) string {
	if price < 1 {
		return "10"
	} else if price < 2 {
		return "5"
	} else {
		return "1"
	}
}
