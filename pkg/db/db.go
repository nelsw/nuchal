package db

import (
	"fmt"
	"github.com/nelsw/nuchal/pkg/util"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	reggol "gorm.io/gorm/logger"
	gol "log"
	"os"
	"time"
)

type Config struct {
	Host string `yml:"host" envconfig:"POSTGRES_HOST"`
	User string `yml:"user" envconfig:"POSTGRES_USER"`
	Pass string `yml:"pass" envconfig:"POSTGRES_PASSWORD"`
	Name string `yml:"name" envconfig:"POSTGRES_DB"`
	Port int    `yml:"port" envconfig:"POSTGRES_PORT"`
}

func (c *Config) dsn() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.Host, c.User, c.Pass, c.Name, c.Port)
}

var cfg *Config

func InitDb() error {

	cfg = new(Config)

	err := util.ConfigFromEnv(cfg)
	if err != nil || cfg.Port == 0 {
		f, err := os.Open("pkg/db/config.yml")
		if err != nil {
			return err
		}
		d := yaml.NewDecoder(f)
		if err := d.Decode(cfg); err != nil {
			return err
		}
	}

	if pg, err := OpenDB(cfg.dsn()); err != nil {
		return err
	} else if sql, err := pg.DB(); err != nil {
		return err
	} else if err := sql.Ping(); err != nil {
		return err
	} else if err := sql.Close(); err != nil {
		return err
	}

	return nil
}

func NewDB() *gorm.DB {
	db, _ := OpenDB(cfg.dsn())
	return db
}

func OpenDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: reggol.New(
			gol.New(os.Stdout, "\r\n", gol.LstdFlags), // io writer
			reggol.Config{
				SlowThreshold:             time.Second,   // Slow SQL threshold
				LogLevel:                  reggol.Silent, // Log level
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,         // Disable color
			},
		),
	})
}
