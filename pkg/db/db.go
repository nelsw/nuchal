package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"nchl/pkg"
	"os"
	"time"
)

type Config struct {
	User, Pass, Name string
	Port             int
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%d", c.User, c.Pass, c.Name, c.Port)
}

var config Config

func init() {
	config = Config{
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
		os.Getenv("POSTGRES_NAME"),
		pkg.Int(os.Getenv("POSTGRES_PORT")),
	}
}

func NewDB() *gorm.DB {
	db, _ := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second,   // Slow SQL threshold
				LogLevel:                  logger.Silent, // Log level
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,         // Disable color
			},
		),
	})
	return db
}
