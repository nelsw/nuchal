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
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"time"
)

var usdRegex = regexp.MustCompile(`^((\w{3,5})(-USD))$`)

type Api struct {
	Key        string `envconfig:"COINBASE_PRO_KEY" yaml:"key"`
	Passphrase string `envconfig:"COINBASE_PRO_PASSPHRASE" yaml:"pass"`
	Secret     string `envconfig:"COINBASE_PRO_SECRET" yaml:"secret"`
	Fees       `yaml:"fees"`
}

type Fees struct {
	Maker float64 `envconfig:"COINBASE_PRO_MAKER_FEE" yaml:"maker"`
	Taker float64 `envconfig:"COINBASE_PRO_TAKER_FEE" yaml:"taker"`
}

func (a *Api) Validate() error {
	if a.Key == "" {
		return errors.New("missing Coinbase Pro API key")
	} else if a.Secret == "" {
		return errors.New("missing Coinbase Pro API secret")
	} else if a.Passphrase == "" {
		return errors.New("missing Coinbase Pro API passphrase")
	}
	return nil
}

func (a *Api) GetClient() *cb.Client {
	return &cb.Client{
		"https://api.pro.coinbase.com",
		a.Secret,
		a.Key,
		a.Passphrase,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}

func (a *Api) GetTime() (*time.Time, error) {
	tme, err := a.GetClient().GetTime()
	if err != nil {
		return nil, err
	}
	t := time.Unix(int64(tme.Epoch), 0)
	return &t, nil
}

func (a *Api) GetFillsByOrderId(id string) (*[]cb.Fill, error) {
	var newChunks, allChunks []cb.Fill
	cursor := a.GetClient().ListFills(cb.ListFillsParams{OrderID: id})
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

func (a *Api) GetFills(productId string) (*[]cb.Fill, error) {

	cursor := a.GetClient().ListFills(cb.ListFillsParams{ProductID: productId})

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

func (a *Api) GetUsdProducts() (*[]cb.Product, error) {

	all, err := a.GetClient().GetProducts()

	var res []cb.Product
	for _, p := range all {
		if usdRegex.MatchString(p.ID) {
			res = append(res, p)
		}
	}

	return &res, err
}

func (a *Api) GetActiveAccounts() (*[]cb.Account, error) {
	allAccounts, err := a.GetClient().GetAccounts()
	if err != nil {
		return nil, err
	}
	var actAccounts []cb.Account
	for _, account := range allAccounts {
		if util.IsZero(account.Balance) && util.IsZero(account.Hold) {
			continue
		}
		actAccounts = append(actAccounts, account)
	}
	return &actAccounts, nil
}

func (a *Api) GetActivePositions() (*map[string]Position, error) {

	accounts, err := a.GetActiveAccounts()
	if err != nil {
		return nil, err
	}

	positions := map[string]Position{}

	for _, account := range *accounts {

		productID := account.Currency + "-USD"

		if account.Currency == "USD" {
			positions[productID] = *NewUsdPosition(account)
			continue
		}

		fills, err := a.GetFills(productID)
		if err != nil {
			return nil, err
		}

		ticker, err := a.GetClient().GetTicker(productID)
		if err != nil {
			return nil, err
		}

		positions[productID] = *NewPosition(account, ticker, *fills)
	}

	return &positions, nil
}

func (a Api) GetOrders(productId string) (*[]cb.Order, error) {

	cursor := a.GetClient().ListOrders(cb.ListOrdersParams{ProductID: productId})

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
func (a *Api) CreateOrder(order *cb.Order, attempt ...int) (*cb.Order, error) {

	r, err := a.GetClient().CreateOrder(order)
	if err == nil {
		return a.GetOrder(r.ID)
	}

	i := util.FirstIntOrZero(attempt)
	if util.IsInsufficientFunds(err) || i > 2 {
		return nil, err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return a.CreateOrder(order, i)
}

// GetOrder is a recursive function that returns an order equal to the given id once it is settled and not pending.
// This function also performs extensive logging given its variable and seriously critical nature.
func (a *Api) GetOrder(id string, attempt ...int) (*cb.Order, error) {

	order, err := a.GetClient().GetOrder(id)

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
		return a.GetOrder(id, i)
	}

	time.Sleep(1 * time.Second)
	return a.GetOrder(id)
}

// CancelOrder is a recursive function that cancels an order equal to the given id.
func (a *Api) CancelOrder(id string, attempt ...int) error {

	err := a.GetClient().CancelOrder(id)
	if err == nil {
		return nil
	}

	i := util.FirstIntOrZero(attempt)
	if i > 10 {
		return err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return a.CancelOrder(id, i)
}

// GetPrice gets the latest ticker price for the given productId. This method does not perform logging as it is executed
// thousands of times per second.
func (a *Api) GetPrice(wsConn *ws.Conn, productID string) (*float64, error) {

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productID}}},
	}); err != nil {
		log.Error().Err(err).Str(util.Currency, productID).Msg(util.Fish + " ... subscribing to websocket")
		return nil, err
	}

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Error().Err(err).Str(util.Currency, productID).Msg(util.Fish + " ... reading from websocket")
			return nil, err
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}

	if receivedMessage.Type != "ticker" {
		err := fmt.Errorf("message type != ticker, %v", receivedMessage)
		log.Error().Err(err).Str(util.Currency, productID).Msg("getting ticker message from websocket")
		return nil, err
	}

	f := util.Float64(receivedMessage.Price)
	return &f, nil
}
