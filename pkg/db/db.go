package db

import (
	"encoding/json"
	"fmt"
	zog "github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gol "gorm.io/gorm/logger"
	"log"
	"nuchal/pkg/util"
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

	config = &Config{
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		util.Int(os.Getenv("POSTGRES_PORT")),
	}

	if config.Port == -1 {
		if file, err := os.Open("pkg/config/database.json"); err == nil {
			_ = json.NewDecoder(file).Decode(&config)
		}
	}

	if db, err := OpenDB(); err != nil {
		zog.Error().Err(err).Send()
	} else if sql, err := db.DB(); err != nil {
		zog.Error().Err(err).Send()
	} else if err := sql.Ping(); err != nil {
		zog.Error().Err(err).Send()
	} else if err := sql.Close(); err != nil {
		zog.Error().Err(err).Send()
	}
}

func NewDB() *gorm.DB {
	db, _ := OpenDB()
	return db
}

func OpenDB() (*gorm.DB, error) {
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
