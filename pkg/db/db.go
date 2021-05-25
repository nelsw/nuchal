package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gol "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func NewDB(dsn string) *gorm.DB {
	db, _ := OpenDB(dsn)
	return db
}

func OpenDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
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
