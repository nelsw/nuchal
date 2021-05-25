package model

import (
	"os"
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

	api, err := NewCoinbaseApi()
	if err != nil {
		return nil, err
	}

	u.CoinbaseApi = api
	return u, nil
}

func (u *User) Validate() error {
	return u.CoinbaseApi.validate()
}
