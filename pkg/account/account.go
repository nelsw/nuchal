package account

import (
	"encoding/json"
	"errors"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nchl/pkg/db"
	"net/http"
	"os"
	"time"
)

type Group struct {
	Users []User `json:"users"`
}

func (c Group) User() *User {
	return &c.Users[0]
}

func (c Group) Client() *cb.Client {
	return c.User().GetClient()
}

func (c *Group) validate() error {
	if c.Users == nil && len(c.Users) < 1 {
		return errors.New("no users found in configuration")
	} else {
		for _, user := range c.Users {
			if err := user.validate(); err != nil {
				return err
			}
		}
		return nil
	}
}

type User struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Secret     string `json:"secret"`
	Enable     bool   `json:"enable"`
}

func (u *User) validate() error {
	if u == (&User{}) {
		return errors.New("account is blank")
	} else if u.Key == "" {
		return errors.New("missing Coinbase Pro API key")
	} else if u.Secret == "" {
		return errors.New("missing Coinbase Pro API secret")
	} else if u.Passphrase == "" {
		return errors.New("missing Coinbase Pro API passphrase")
	} else {
		return nil
	}
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

type Portfolio struct {
	Username, Value string
	Positions       []Status
}

type Status struct {
	ProductId, Value string
	Balance          float64
}

func NewGroup() (*Group, error) {

	log.Info().Msg("creating account group")

	c := new(Group)

	if file, err := os.Open("assets/users.json"); err == nil {
		log.Info().Msg("found a users config file")
		if err := json.NewDecoder(file).Decode(&c); err != nil {
			log.Warn().Err(err).Msg("unable to decode users.json")
		} else if err := c.validate(); err != nil {
			log.Warn().Err(err)
		} else {
			return c, nil
		}
	}

	db.NewDB().Find(&c.Users)
	if err := c.validate(); err != nil {
		log.Warn().Err(err)
	} else {
		return c, nil
	}

	c.Users = append(c.Users, User{
		os.Getenv("name"),
		os.Getenv("key"),
		os.Getenv("pass"),
		os.Getenv("secret"),
		true,
	})
	if err := c.validate(); err != nil {
		log.Error().Err(err)
		return nil, err
	}

	log.Info().Msg("created account group")

	return c, nil
}
