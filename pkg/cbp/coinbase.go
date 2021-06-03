package cbp

import (
	"errors"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"net/http"
	"strings"
	"time"
)

type Api struct {
	Key        string `envconfig:"COINBASE_PRO_KEY" json:"key" yaml:"key"`
	Passphrase string `envconfig:"COINBASE_PRO_PASSPHRASE" json:"passphrase" yaml:"pass"`
	Secret     string `envconfig:"COINBASE_PRO_SECRET" json:"secret" yaml:"secret"`
	Fees       `yaml:"fees"`
}

type Fees struct {
	Maker float64 `envconfig:"COINBASE_PRO_MAKER_FEE" json:"maker_fee" yaml:"maker" default:"0.005"`
	Taker float64 `envconfig:"COINBASE_PRO_TAKER_FEE" json:"taker_fee" yaml:"taker" default:"0.005"`
}

func NewApi() (*Api, error) {

	c := new(Api)

	if err := util.ConfigFromEnv(c); err != nil {
		return nil, err
	} else if err := c.validate(); err != nil {
		if err := util.ConfigFromYml(c); err == nil {
			if err := c.validate(); err == nil {
				return c, nil
			}
		}
	}

	return c, nil
}

func (a *Api) validate() error {
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
		if strings.Contains(p.ID, "-USD") {
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
