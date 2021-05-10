package pkg

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gol "log"
	"os"
	"time"
)

var db *gorm.DB

const dsn = "host=localhost user=postgres password=somePassword dbname=nuchal port=5432"

func init() {
	var err error
	if db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
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
