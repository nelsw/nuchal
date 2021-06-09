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

package cbp

import (
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
	"regexp"
	"time"
)

type Config struct {
	Api struct {
		Key        string `envconfig:"COINBASE_PRO_KEY" yaml:"key"`
		Passphrase string `envconfig:"COINBASE_PRO_PASSPHRASE" yaml:"pass"`
		Secret     string `envconfig:"COINBASE_PRO_SECRET" yaml:"secret"`
		Fees       struct {
			Maker float64 `envconfig:"COINBASE_PRO_MAKER_FEE" yaml:"maker"`
			Taker float64 `envconfig:"COINBASE_PRO_TAKER_FEE" yaml:"taker"`
		} `yaml:"fees"`
	} `yaml:"cbp"`
}

var (
	cfg      *Config
	client   *cb.Client
	products = map[string]cb.Product{}
	usdRegex = regexp.MustCompile(`^((\w{3,5})(-USD))$`)
)

func Init(name string) (*time.Time, error) {

	var err error

	cfg = new(Config)

	if err = envconfig.Process("", cfg); err == nil {
		err = validate()
	}

	if err != nil {
		var f *os.File
		if f, err = os.Open(name); err == nil {
			if err = yaml.NewDecoder(f).Decode(cfg); err == nil {
				err = validate()
			}
		}
	}

	baseUrl := "https://api.pro.coinbase.com"
	if err != nil {
		baseUrl = "https://api-public.sandbox.pro.coinbase.com"
	}

	client = &cb.Client{
		baseUrl,
		cfg.Api.Secret,
		cfg.Api.Key,
		cfg.Api.Passphrase,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}

	var allProducts []cb.Product
	allProducts, err = client.GetProducts()
	if err != nil {
		return nil, err
	}

	for _, product := range allProducts {
		if product.BaseCurrency == "DAI" ||
			product.BaseCurrency == "USDT" ||
			product.BaseMinSize == "" ||
			product.QuoteIncrement == "" ||
			!usdRegex.MatchString(product.ID) {
			continue
		}
		products[product.ID] = product
	}

	if cfg.Api.Fees.Maker == 0 {
		cfg.Api.Fees.Maker = .005
	}
	if cfg.Api.Fees.Taker == 0 {
		cfg.Api.Fees.Taker = .005
	}

	var tme cb.ServerTime
	if tme, err = client.GetTime(); err != nil {
		return nil, err
	}

	sec := int64(tme.Epoch)
	now := time.Unix(sec, 0)

	return &now, nil
}

func validate() error {
	if cfg.Api.Key == "" {
		return errors.New("missing Coinbase Pro API key")
	} else if cfg.Api.Secret == "" {
		return errors.New("missing Coinbase Pro API secret")
	} else if cfg.Api.Passphrase == "" {
		return errors.New("missing Coinbase Pro API passphrase")
	}
	return nil
}

func Maker() float64 {
	return cfg.Api.Fees.Maker
}

func Taker() float64 {
	return cfg.Api.Fees.Taker
}

func GetAllProductIDs() []string {
	var productIDs []string
	for productID := range products {
		productIDs = append(productIDs, productID)
	}
	return productIDs
}

func GetHistoricRates(productID string, start, end time.Time) ([]cb.HistoricRate, error) {
	return client.GetHistoricRates(productID, cb.GetHistoricRatesParams{start, end, 60})
}

func GetFills(productID string) (*[]cb.Fill, error) {

	cursor := client.ListFills(cb.ListFillsParams{ProductID: productID})

	var newChunks, allChunks []cb.Fill
	for cursor.HasMore {

		if err := cursor.NextPage(&newChunks); err != nil {
			return nil, err
		}

		for _, chunk := range newChunks {
			allChunks = append(allChunks, chunk)
		}
	}

	return &allChunks, nil
}

func GetActivePositions() (map[string]Position, error) {

	accounts, err := client.GetAccounts()
	if err != nil {
		return nil, err
	}

	positions := map[string]Position{}

	for _, account := range accounts {

		if util.IsZero(account.Balance) && util.IsZero(account.Hold) {
			continue
		}

		productID := account.Currency + "-USD"

		if account.Currency == "USD" {
			positions[productID] = *NewUsdPosition(account)
			continue
		}

		fills, err := GetFills(productID)
		if err != nil {
			return nil, err
		}

		ticker, err := client.GetTicker(productID)
		if err != nil {
			return nil, err
		}

		positions[productID] = *NewPosition(account, ticker, *fills)
	}

	return positions, nil
}

// GetTradingPositions returns a map of trading positions.
func GetTradingPositions() (map[string]Position, error) {

	positions, err := GetActivePositions()
	if err != nil {
		return nil, err
	}

	result := map[string]Position{}
	for productID, position := range positions {
		if position.Currency == "USD" || position.Balance() == position.Hold() {
			continue
		}
		result[productID] = position
	}

	return result, nil
}

func GetOrders(productID string) (*[]cb.Order, error) {

	cursor := client.ListOrders(cb.ListOrdersParams{ProductID: productID})

	var newChunks, allChunks []cb.Order
	for cursor.HasMore {

		if err := cursor.NextPage(&newChunks); err != nil {
			return nil, err
		}

		for _, chunk := range newChunks {
			allChunks = append(allChunks, chunk)
		}
	}

	return &allChunks, nil
}

// CreateOrder creates an order on Coinbase and returns the order once it is no longer pending and has settled.
// Given that there are many different types of orders that can be created in many different scenarios, it is the
// responsibility of the method calling this function to perform logging.
func CreateOrder(order *cb.Order, attempt ...int) (*cb.Order, error) {

	r, err := client.CreateOrder(order)
	if err == nil {
		return GetOrder(r.ID)
	}

	i := util.FirstIntOrZero(attempt)
	if util.IsInsufficientFunds(err) || i > 2 {
		return nil, err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return CreateOrder(order, i)
}

// GetOrder is a recursive function that returns an order equal to the given id once it is settled and not pending.
// This function also performs extensive logging given its variable and seriously critical nature.
func GetOrder(id string, attempt ...int) (*cb.Order, error) {

	order, err := client.GetOrder(id)

	if err == nil && order.Status != "pending" {
		return &order, nil
	}

	if err != nil {

		i := util.FirstIntOrZero(attempt)
		if i > 10 {
			return nil, err
		}

		i++
		time.Sleep(time.Duration(i*3) * time.Second)
		return GetOrder(id, i)
	}

	time.Sleep(1 * time.Second)
	return GetOrder(id)
}

// CancelOrder is a recursive function that cancels an order equal to the given id.
func CancelOrder(id string, attempt ...int) error {

	err := client.CancelOrder(id)
	if err == nil {
		return nil
	}

	i := util.FirstIntOrZero(attempt)
	if i > 10 {
		return err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return CancelOrder(id, i)
}

// GetPrice gets the latest ticker price for the given productID. This method does not perform logging as it is executed
// thousands of times per second.
func GetPrice(wsConn *ws.Conn, productID string) (*float64, error) {

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productID}}},
	}); err != nil {
		return nil, err
	}

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			return nil, err
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}

	if receivedMessage.Type != "ticker" {
		err := fmt.Errorf("message type != ticker, %v", receivedMessage)
		return nil, err
	}

	f := util.Float64(receivedMessage.Price)
	return &f, nil
}

func GetRate(productID string) (*Rate, error) {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		return nil, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
		}
	}(wsConn)

	end := time.Now().Add(time.Minute)

	var low, high, open, vol float64
	for {

		price, err := GetPrice(wsConn, productID)
		if err != nil {
			return nil, err
		}

		vol++

		if low == 0 {
			low = *price
			high = *price
			open = *price
		} else if high < *price {
			high = *price
		} else if low > *price {
			low = *price
		}

		if now := time.Now(); now.After(end) {
			return &Rate{
				now.UnixNano(),
				productID,
				cb.HistoricRate{now, low, high, open, *price, vol},
			}, nil
		}
	}
}

func GetTickerPrice(productID string) (*float64, error) {
	ticker, err := client.GetTicker(productID)
	if err != nil {
		return nil, err
	}
	price := util.Float64(ticker.Price)
	return &price, nil
}
