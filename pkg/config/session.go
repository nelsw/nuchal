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
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
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

	Started time.Time
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

func (s Session) ProductIds() *[]string {
	var productIds []string
	for productId, _ := range s.Products {
		productIds = append(productIds, productId)
	}
	sort.Strings(productIds)
	return &productIds
}

func (s *Session) User() string {
	return os.Getenv("USER")
}

// NewSession reads configuration from environment variables and validates it
func NewSession(usd []string, size, gain, loss, delta float64) (*Session, error) {

	util.PrintlnBanner()

	session := new(Session)

	log.Info().Msg(util.Fish + " . ")
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " ... hello " + session.User())
	log.Info().Msg(util.Fish + " .. ")
	log.Info().Msg(util.Fish + " . ")

	err := util.ConfigFromYml(session)
	if err != nil {
		// either the file didn't exist or wasn't properly formatted
		err = util.ConfigFromEnv(session)
		if err != nil {
			return nil, err
		}
	}

	// Lets confirm our API credentials are correct
	if err := session.Api.Validate(); err != nil {
		return nil, err
	}

	// Set a "start time" for the session
	if tme, err := session.GetTime(); err != nil {
		return nil, err
	} else {
		session.Started = *tme
	}

	// No other place to really put this
	if session.Port == 0 {
		session.Port = 8080
	}

	// While trade and report commands do not requiring the database,
	// simulations do, which should occur before trading & reporting.
	if err := db.InitDb(); err != nil {
		return nil, err
	}

	// Initiating products will fetch the most recent list of
	// cryptocurrencies and apply patterns to each product.
	allUsdProducts, err := session.Api.GetUsdProducts()
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
		mm[pattern.Id] = cbp.Product{product, pattern}
	}
	session.Products = mm

	log.Debug().Interface("session", session).Send()

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

	return session, nil
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

func (s *Session) GetCurrentPrice(productId string) (*float64, error) {
	ticker, err := s.GetClient().GetTicker(productId)
	if err != nil {
		return nil, err
	}
	price := util.Float64(ticker.Price)
	return &price, nil
}

// GetPrice gets the latest ticker price for the given productId. This method does not perform logging as it is executed
// thousands of times per second.
func (s *Session) GetPrice(wsConn *ws.Conn, productId string) (*float64, error) {

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productId}}},
	}); err != nil {
		log.Error().Err(err).Str(util.Currency, productId).Msg(util.Fish + " ... subscribing to websocket")
		return nil, err
	}

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Error().Err(err).Str(util.Currency, productId).Msg(util.Fish + " ... reading from websocket")
			return nil, err
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}

	if receivedMessage.Type != "ticker" {
		err := errors.New(fmt.Sprintf("message type != ticker, %v", receivedMessage))
		log.Error().Err(err).Str(util.Currency, productId).Msg("getting ticker message from websocket")
		return nil, err
	}

	f := util.Float64(receivedMessage.Price)
	return &f, nil
}
