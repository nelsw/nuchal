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

package config

import (
	"bufio"
	"fmt"
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

// Session of the application - only active while the executable is run within the defined period range.
type Session struct {

	// Port is the port where nuchal will serve (simulation) html files.
	Port int `envconfig:"PORT"`

	// Period is a range of time representing when to start and stop executing the trade command.
	Period `yaml:"period"`

	*cbp.Api `yaml:"cbp"`

	Products map[string]cbp.Product
}

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if util.IsEnvVarTrue("DEBUG") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("***%s****", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}
}

func (s *Session) User() string {
	return os.Getenv("USER")
}

// NewSession reads configuration from environment variables and validates it
func NewSession(usd []string, size, gain, loss, delta float64) (*Session, error) {

	util.PrintlnBanner()

	c := new(Session)

	log.Info().Msg(util.Fish + " . ")
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " ... hello " + c.User())
	log.Info().Msg(util.Fish + " .. ")
	log.Debug().Msg("configure session")

	err := util.ConfigFromYml(c)
	if err != nil {
		// either the file didn't exist or wasn't properly formatted
		err = util.ConfigFromEnv(c)
		if err != nil {
			return nil, err
		}
	}

	// Lets confirm our API credentials are correct
	if err := c.Api.Validate(); err != nil {
		return nil, err
	}

	// No other place to reall put this
	if c.Port == 0 {
		c.Port = 8080
	}

	// While trade and report commands do not requiring the database,
	// simulations do, which should occur before trading & reporting.
	if err := db.InitDb(); err != nil {
		return nil, err
	}

	// Initiating products will fetch the most recent list of
	// cryptocurrencies and apply patterns to each product.
	allUsdProducts, err := c.Api.GetUsdProducts()

	if err != nil {
		return nil, err
	}
	allProductsMap := map[string]cb.Product{}
	for _, a := range *allUsdProducts {
		if a.BaseCurrency == "DAI" || a.BaseCurrency == "USDT" {
			continue
		}
		allProductsMap[a.ID] = a
	}

	productMap := map[string]cb.Product{}
	if usd != nil {
		for _, currency := range usd {
			productId := currency + "-USD"
			productMap[productId] = allProductsMap[productId]
		}
	} else {
		productMap = allProductsMap
	}

	var wrapper struct {
		Patterns []cbp.Pattern `yaml:"patterns"`
	}

	if err := util.ConfigFromYml(&wrapper); err != nil {
		return nil, err
	}

	if len(wrapper.Patterns) < 1 {
		for _, product := range *allUsdProducts {
			pattern := new(cbp.Pattern)
			pattern.Id = product.ID
			wrapper.Patterns = append(wrapper.Patterns, *pattern)
		}
	}

	mm := map[string]cbp.Product{}
	for _, pattern := range wrapper.Patterns {
		product := productMap[pattern.Id]
		if product.BaseMinSize == "" || product.QuoteIncrement == "" {
			continue
		}
		if pattern.Size == 0 {
			pattern.Size = util.Float64(product.BaseMinSize) * size
		}
		if pattern.Gain == 0 {
			pattern.Gain = gain
		}
		if pattern.Loss == 0 {
			pattern.Loss = loss
		}
		if pattern.Delta == 0 {
			pattern.Delta = delta
		}
		mm[pattern.Id] = cbp.Product{product, pattern}
	}
	c.Products = mm

	log.Debug().Interface("session", c).Msg("configure")
	log.Info().Msg(util.Fish + " ... everything seems to be in order")
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " . ")

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if line == "exit" {
				exit()
			} else {
				log.Info().Msg(util.Fish + " ... I'm not familiar with ")
			}
		}
		if err := scanner.Err(); err != nil {
			log.Error().Err(err).Send()
			panic(err)
		}
	}()

	return c, nil
}

func exit() {
	log.Info().Msg(util.Fish + " .")
	log.Info().Msg(util.Fish + " ..")
	log.Info().Msg(util.Fish + " ...")
	log.Info().Msg(util.Fish + " ... goodbye")
	log.Info().Msg(util.Fish + " ...")
	log.Info().Msg(util.Fish + " ..")
	log.Info().Msg(util.Fish + " .")
	os.Exit(0)
}

func (s *Session) GetTradingPositions() (*[]cbp.Position, error) {

	positions, err := s.GetActivePositions()
	if err != nil {
		return nil, err
	}

	var result []cbp.Position
	for _, position := range *positions {
		if position.Currency == "USD-USD" || position.Balance() == position.Hold() {
			continue
		}
		result = append(result, position)
	}

	return &result, nil
}

func (s *Session) GetActivePositions() (*[]cbp.Position, error) {

	positions, err := s.Api.GetActivePositions()
	if err != nil {
		return nil, err
	}

	var result []cbp.Position
	for _, position := range *positions {
		if position.Currency != "USD" {
			position.Pattern = s.Products[position.ProductId()].Pattern
		}
		result = append(result, position)
	}

	return &result, nil
}
