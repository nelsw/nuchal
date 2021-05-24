package model

import (
	"errors"
	"github.com/kelseyhightower/envconfig"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"net/http"
	"time"
)

type CoinbaseApi struct {
	Key        string  `envconfig:"COINBASE_PRO_KEY" json:"key"`
	Passphrase string  `envconfig:"COINBASE_PRO_PASSPHRASE" json:"passphrase"`
	Secret     string  `envconfig:"COINBASE_PRO_SECRET" json:"secret"`
	MakerFee   float64 `envconfig:"COINBASE_PRO_MAKER_FEE" default:"0.005" json:"maker_fee"`
	TakerFee   float64 `envconfig:"COINBASE_PRO_TAKER_FEE" default:"0.005" json:"taker_fee"`
}

func NewCoinbaseApi() (*CoinbaseApi, error) {
	c := new(CoinbaseApi)
	err := envconfig.Process("", c)
	return c, err
}

func (c *CoinbaseApi) validate() error {
	if c.Key == "" {
		return errors.New("missing Coinbase Pro API key")
	} else if c.Secret == "" {
		return errors.New("missing Coinbase Pro API secret")
	} else if c.Passphrase == "" {
		return errors.New("missing Coinbase Pro API passphrase")
	}
	return nil
}

func (c *CoinbaseApi) GetClient() *cb.Client {
	return &cb.Client{
		"https://api.pro.coinbase.com",
		c.Secret,
		c.Key,
		c.Passphrase,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}
