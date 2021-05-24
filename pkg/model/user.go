package model

import (
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"net/http"
	"os"
	"time"
)

type User struct {
	Name   string `json:"name"`
	Enable bool   `json:"enable"`
	*CoinbaseApi
}

func NewUser() (*User, error) {

	u := new(User)
	u.Enable = true

	u.Name = os.Getenv("USER")
	if u.Name == "" {
		u.Name = util.GuestName
	}

	api, err := NewCoinbaseApi()
	if err != nil {
		return nil, err
	}

	u.CoinbaseApi = api
	return u, nil
}

func (u *User) validate() error {
	return u.CoinbaseApi.validate()
}

func (u *User) GetClient() *cb.Client {
	return &cb.Client{
		"https://api.pro.coinbase.com",
		u.Secret,
		u.Key,
		u.Passphrase,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}
