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
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
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

func (a *Api) GetActivePositions() (*[]Position, error) {

	accounts, err := a.GetActiveAccounts()
	if err != nil {
		return nil, err
	}

	var positions []Position

	for _, account := range *accounts {

		if account.Currency == "USD" {
			positions = append(positions, *NewUsdPosition(account))
			continue
		}

		productId := account.Currency + "-USD"

		fills, err := a.GetFills(productId)
		if err != nil {
			return nil, err
		}

		ticker, err := a.GetClient().GetTicker(productId)
		if err != nil {
			return nil, err
		}

		positions = append(positions, *NewPosition(account, ticker, *fills))
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
