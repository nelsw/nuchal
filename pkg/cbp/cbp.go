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
	"github.com/kelseyhightower/envconfig"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
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

	if cfg.Api.Fees.Maker == 0 {
		cfg.Api.Fees.Maker = .005
	}
	if cfg.Api.Fees.Taker == 0 {
		cfg.Api.Fees.Taker = .005
	}

	if allProducts, err := client.GetProducts(); err != nil {
		return nil, err
	} else {
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

func Client() *cb.Client {
	return client
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

func GetProduct(productID string) cb.Product {
	return products[productID]
}
