package db

import (
	"encoding/json"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gol "log"
	"os"
	"time"
)

type Configuration struct {
	DSN string `json:"dsn"`
}

var Client *gorm.DB
var dbConfig Configuration

func init() {
	if file, err := os.Open("./.app/db/config.json"); err != nil {
		panic(err)
	} else if err = json.NewDecoder(file).Decode(&dbConfig); err != nil {
		panic(err)
	}
	var err error
	if Client, err = gorm.Open(postgres.Open(dbConfig.DSN), &gorm.Config{
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
}
