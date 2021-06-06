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
	"sort"
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

	Patterns []cbp.Pattern `yaml:"patterns"`

	*time.Time
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

// ProductIds returns a product ID array in alphabetical order.
func (s Session) ProductIds() *[]string {
	var productIds []string
	for _, product := range s.Products {
		productIds = append(productIds, product.ID)
	}
	sort.Strings(productIds)
	return &productIds
}

func hello() {
	log.Info().Msg(util.Fish + " . ")
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " ... hello " + os.Getenv("USER"))
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " . ")
}

// NewSession reads configuration from environment variables and validates it
func NewSession(cfg string, usd []string, size, gain, loss, delta float64) (*Session, error) {

	util.PrintlnBanner()

	session := new(Session)

	hello()

	if err := util.ConfigFromYml(session, cfg); err != nil {
		// either the file didn't exist or wasn't properly formatted
		if err = util.ConfigFromEnv(session); err != nil {
			return nil, err
		}
	}

	// Lets confirm our API credentials are correct
	if err := session.Api.Validate(); err != nil {
		return nil, err
	}

	// While trade and report commands do not requiring the database,
	// simulations do, which should occur before trading & reporting.
	if err := db.InitDb(); err != nil {
		return nil, err
	}

	// Set a "start time" for the session
	tme, err := session.GetTime()
	if err != nil {
		return nil, err
	}
	session.Time = tme

	// No other place to really put this
	if session.Port == 0 {
		session.Port = 8080
	}

	// Initiating products will fetch the most recent list of
	// cryptocurrencies and apply patterns to each product.
	allUsdProducts, err := session.Api.GetUsdProducts()
	if err != nil {
		return nil, err
	}

	allProductsMap := map[string]cb.Product{}
	for _, product := range *allUsdProducts {
		if product.BaseCurrency == "DAI" || product.BaseCurrency == "USDT" {
			continue
		}
		allProductsMap[product.ID] = product
	}

	session.Products = map[string]cbp.Product{}
	if len(session.Patterns) < 1 {
		for productID, product := range allProductsMap {
			session.Products[productID] = cbp.Product{product, *cbp.NewPattern(size, gain, loss, delta)}
		}
	}

	for _, pattern := range session.Patterns {
		product := allProductsMap[pattern.Id]
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
		session.Products[pattern.Id] = cbp.Product{product, pattern}
	}

	log.Debug().Interface("session", session).Send()

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if line == "exit" {
				goodbye()
			} else {
				log.Info().Msg(util.Fish + " ... I'm not familiar with ")
			}
		}
		if err := scanner.Err(); err != nil {
			log.Error().Err(err).Send()
			panic(err)
		}
	}()

	return session, nil
}

func goodbye() {
	log.Info().Msg(util.Fish + " .")
	log.Info().Msg(util.Fish + " ..")
	log.Info().Msg(util.Fish + " ...")
	log.Info().Msg(util.Fish + " ... goodbye")
	log.Info().Msg(util.Fish + " ...")
	log.Info().Msg(util.Fish + " ..")
	log.Info().Msg(util.Fish + " .")
	os.Exit(0)
}

// GetTradingPositions returns a map of trading positions.
func (s *Session) GetTradingPositions() (map[string]cbp.Position, error) {

	positions, err := s.GetActivePositions()
	if err != nil {
		return nil, err
	}

	result := map[string]cbp.Position{}
	for _, position := range *positions {
		if position.Currency == "USD" || position.Balance() == position.Hold() {
			continue
		}
		result[position.ProductId()] = position
	}

	return result, nil
}

// GetActivePositions returns an array of cbp.Position structs.
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
