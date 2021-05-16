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
		var enabledUsers []User
		for _, user := range c.Users {
			if err := user.validate(); err != nil {
				return err
			}
			if user.Enable {
				enabledUsers = append(enabledUsers, user)
			}
		}
		c.Users = enabledUsers
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

	g := new(Group)

	var err error
	if err = loadFromJson(g); err != nil {
		log.Warn().Err(err).Msg("user load from json failed")
		if err = loadFromDatabase(g); err != nil {
			log.Warn().Err(err).Msg("user load from database failed")
			if err = loadFromEnvironment(g); err != nil {
				log.Warn().Err(err).Msg("user load from environment failed")
			} else {
				log.Info().Msg("created account group")
			}
		}
	}

	if err == nil {
		log.Info().Msgf("created account group [%v]", g)
	} else {
		log.Error().Err(err).Msg("error creating account group")
	}

	return g, err
}

func loadFromJson(g *Group) error {
	if file, err := os.Open("assets/users.json"); err != nil {
		// no user json file, not worth the log space
		return err
	} else if err := json.NewDecoder(file).Decode(&g); err != nil {
		log.Warn().Err(err).Msg("unable to decode users.json")
		return err
	} else if err := g.validate(); err != nil {
		log.Warn().Err(err).Msg("user json was invalid")
		return err
	} else {
		log.Info().Msg("created account group from user json")
		return nil
	}
}

func loadFromDatabase(g *Group) error {
	db.NewDB().Find(&g.Users)
	if err := g.validate(); err != nil {
		log.Warn().Err(err)
		return err
	}
	log.Info().Msg("created account group from database")
	return nil
}

func loadFromEnvironment(g *Group) error {
	g.Users = append(g.Users, User{
		os.Getenv("name"),
		os.Getenv("key"),
		os.Getenv("pass"),
		os.Getenv("secret"),
		true,
	})
	if err := g.validate(); err != nil {
		log.Error().Err(err).Msg("user environment variables were invalid")
		return err
	}
	log.Info().Msg("created account group from environment variables")
	return nil
}
