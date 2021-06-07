/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package db

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
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

	err := envconfig.Process("", cfg)
	if err != nil || cfg.Port == 0 {
		cfg.Host = "localhost"
		cfg.User = "postgres"
		cfg.Name = "nuchal"
		cfg.Pass = "somePassword"
		cfg.Port = 5432
	}

	if pg, err := openDB(cfg.dsn()); err != nil {
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
	db, _ := openDB(cfg.dsn())
	return db
}

func openDB(dsn string) (*gorm.DB, error) {
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
