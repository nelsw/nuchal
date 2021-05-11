package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gol "log"
	"nchl/pkg/config"
	"os"
	"time"
)

var Client *gorm.DB

func init() {
	var err error
	if Client, err = gorm.Open(postgres.Open(config.DatabaseUrl()), &gorm.Config{
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
