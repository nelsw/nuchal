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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Session of the application - only active while the executable is run within the defined period range.
type Session struct {

	// Period is a range of time representing when to start and stop executing the trade command.
	Period `yaml:"period"`

	*cbp.Api `yaml:"cbp"`

	products map[string]cbp.Product

	Patterns []cbp.Pattern `yaml:"patterns"`

	UsdSelections map[string]string
}

func (s Session) GetProduct(productID string) cbp.Product {
	return s.products[productID]
}

func (s Session) SimPort() int {

	port := os.Getenv("PORT")
	if len(port) < 4 {
		port = "8080"
	}

	prt, err := strconv.Atoi(port)
	if err != nil {
		log.Error().Err(err).Send()
		prt = 8080
	}

	return prt
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

// ProductIDs returns a product ID array in alphabetical order.
func (s *Session) ProductIDs() *[]string {
	var productIDs []string
	for productID, product := range s.products {
		if _, ok := s.UsdSelections[productID]; ok {
			productIDs = append(productIDs, product.ID)
		}
	}
	sort.Strings(productIDs)
	return &productIDs
}

// NewSession reads configuration from environment variables and validates it
func NewSession(cfg string, usd []string, size, gain, loss, delta float64, debug ...bool) (*Session, error) {

	fmt.Println(util.Banner)
	log.Info().Msg(util.Fish + " . ")
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " ... hello " + os.Getenv("USER"))
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " . ")

	if debug != nil && len(debug) > 0 && debug[0] {
		if err := os.Setenv("DEBUG", "true"); err != nil {
			return nil, err
		}
	}

	session := new(Session)

	// can we connect to coinbase?
	if now, err := cbp.Init(cfg); err != nil {
		return nil, err
	} else {
		session.started = now
		log.Info().Msg(util.Fish + " ... coinbase validated")
		log.Info().Msgf("%s ... cryptocurrency products found [%d]", util.Fish, len(cbp.GetAllProductIDs()))
	}

	// is the database established?
	if err := db.Init(); err != nil {
		return nil, err
	}
	log.Info().Msg(util.Fish + " ... database connected")

	if f, err := os.Open(cfg); err == nil {
		if err = yaml.NewDecoder(f).Decode(session); err != nil {
			return nil, err
		}
	}

	// Map all patterns by product ID
	patterns := map[string]cbp.Pattern{}
	for _, pattern := range session.Patterns {
		patterns[pattern.Id] = pattern
	}
	log.Info().Msgf("%s ... patterns configurations found [%d]", util.Fish, len(session.products))

	// If no product patterns have been configured
	if len(patterns) < 1 {
		// create new patterns for every Usd product
		for _, productID := range cbp.GetAllProductIDs() {
			patterns[productID] = *cbp.NewPattern(productID, size, gain, loss, delta)
		}
	}

	// now we have a map of coinbase products and nuchal patterns, make the session products map
	session.products = map[string]cbp.Product{}
	for productID, pattern := range patterns {
		product := cbp.GetProduct(productID)
		pattern.InitPattern(size, gain, loss, delta)
		session.products[productID] = cbp.Product{product, pattern}
	}
	log.Info().Msgf("%s ... product patterns configured [%d]", util.Fish, len(session.products))

	session.UsdSelections = map[string]string{}
	if len(usd) > 0 {
		for _, currency := range usd {
			productID := currency + "-USD"
			if _, ok := session.products[productID]; ok {
				session.UsdSelections[productID] = fmt.Sprintf("%5s", currency)
			}
		}
		log.Info().Msgf("%s ... USD product selections [%d]", util.Fish, len(usd))
	} else {
		for productID := range session.products {
			currency := strings.Split(productID, "-")[0]
			session.UsdSelections[productID] = fmt.Sprintf("%5s", currency)
		}
		log.Info().Msgf("%s ... USD product selections [%d]", util.Fish, len(session.products))
	}

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if line == "exit" {
				log.Info().Msg(util.Fish + " .")
				log.Info().Msg(util.Fish + " ..")
				log.Info().Msg(util.Fish + " ...")
				log.Info().Msg(util.Fish + " ... goodbye")
				log.Info().Msg(util.Fish + " ...")
				log.Info().Msg(util.Fish + " ..")
				log.Info().Msg(util.Fish + " .")
				os.Exit(0)
			} else {
				log.Info().Msgf("%s %s I'm not familiar with \"%s\"", util.Fish, util.Break, line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Error().Err(err).Send()
			panic(err)
		}
	}()

	return session, nil
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
func (s *Session) GetActivePositions() (*map[string]cbp.Position, error) {

	positions, err := s.Api.GetActivePositions()
	if err != nil {
		return nil, err
	}

	result := map[string]cbp.Position{}
	for productID, position := range *positions {
		product := s.products[productID]
		if position.Product == (cbp.Product{}) {
			position.Product = product
		}
		if position.Currency != "USD" {
			position.Pattern = product.Pattern
		}
		result[productID] = position
	}

	return &result, nil
}
