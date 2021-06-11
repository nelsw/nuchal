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
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	reggol "gorm.io/gorm/logger"
	gol "log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host string `envconfig:"POSTGRES_HOST"`
	User string `envconfig:"POSTGRES_USER"`
	Pass string `envconfig:"POSTGRES_PASSWORD"`
	Name string `envconfig:"POSTGRES_DB"`
	Port int    `envconfig:"POSTGRES_PORT"`
}

func (c *Config) dsn() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.Host, c.User, c.Pass, c.Name, c.Port)
}

var cfg *Config

func Init() error {

	if bytes, err := exec.Command("/bin/sh", "-c", "docker ps --format '{{.Names}}'").Output(); err != nil {
		return err
	} else {
		name := strings.TrimSpace(string(bytes))
		if name != "nuchal_db" {
			if _, err := exec.Command("/bin/sh", "-c", "docker compose -p nuchal up -d").Output(); err != nil {
				return err
			}
			time.Sleep(time.Second * 5) // wait a few seconds to allow the docker composition to spin up
		}
	}

	cfg = new(Config)

	_ = envconfig.Process("", cfg)
	if err := cfg.validate(); err == nil {
		return nil
	}

	if envs, err := godotenv.Read(".env"); err == nil {
		if port, err := strconv.Atoi(envs["POSTGRES_PORT"]); err == nil {
			cfg.Host = envs["POSTGRES_HOST"]
			cfg.User = envs["POSTGRES_USER"]
			cfg.Name = envs["POSTGRES_DB"]
			cfg.Pass = envs["POSTGRES_PASSWORD"]
			cfg.Port = port
			if err = cfg.validate(); err == nil {
				return nil
			}
		}
	}

	cfg.Host = "localhost"
	cfg.User = "postgres"
	cfg.Name = "nuchal"
	cfg.Port = 5432
	cfg.Pass = "somePassword"

	return nil
}

func (c Config) validate() error {

	if pg, err := openDB(c.dsn()); err != nil {
		return err
	} else if sql, err := pg.DB(); err != nil {
		return err
	} else if err := sql.Ping(); err != nil {
		return err
	} else if err := sql.Close(); err != nil {
		return err
	} else if err := pg.AutoMigrate(cbp.Rate{}); err != nil {
		return err
	} else if err := pg.AutoMigrate(cbp.Product{}); err != nil {
		return err
	}

	return nil
}

func NewDB(vv ...interface{}) *gorm.DB {

	db, err := openDB(cfg.dsn())
	if err == nil && vv != nil && len(vv) > 0 {
		for _, v := range vv {
			if err = db.AutoMigrate(v); err != nil {
				break
			}
		}
	}

	if err != nil {
		log.Debug().Err(err).Send()
	}

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
