package db

import (
	"fmt"
	zog "github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gol "gorm.io/gorm/logger"
	"log"
	"nchl/pkg/util"
	"os"
	"time"
)

type Config struct {
	Host, User, Pass, Name string
	Port                   int
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.Host, c.User, c.Pass, c.Name, c.Port)
}

var config *Config

func init() {
	zog.Info().Msg("initializing db")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "host.docker.internal"
	}
	config = &Config{
		host,
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		util.Int(os.Getenv("POSTGRES_PORT")),
	}
	if db, err := openDB(); err != nil {
		zog.Error().Err(err)
		panic(err)
	} else if sql, err := db.DB(); err != nil {
		zog.Error().Err(err)
		panic(err)
	} else if err := sql.Ping(); err != nil {
		zog.Error().Err(err)
		panic(err)
	} else if err := sql.Close(); err != nil {
		zog.Error().Err(err)
		panic(err)
	} else {
		zog.Info().Msg("initialized db")
	}
}

func NewDB() *gorm.DB {
	db, err := openDB()
	if err != nil {
		zog.Error().Err(err).Msg("error opening DB!")
	}
	return db
}

func openDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: gol.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			gol.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  gol.Silent,  // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,       // Disable color
			},
		),
	})
}
