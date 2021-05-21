package account

import (
	"errors"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"net/http"
	"time"
)

type User struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Secret     string `json:"secret"`
	Enable     bool   `json:"enable"`
}

func (u *User) validate() error {
	if u.Key == "" {
		return errors.New("missing Coinbase Pro API key")
	} else if u.Secret == "" {
		return errors.New("missing Coinbase Pro API secret")
	} else if u.Passphrase == "" {
		return errors.New("missing Coinbase Pro API passphrase")
	}
	return nil
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
