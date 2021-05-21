package account

import (
	"encoding/json"
	"errors"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal/pkg/db"
	"os"
	"strings"
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

func NewGroup() (*Group, error) {

	log.Info().Msg("creating account group")

	g := new(Group)

	var err error
	if err = loadFromJson(g); err != nil {
		log.Warn().Err(err).Msg("group load from json failed")
		if err = loadFromDatabase(g); err != nil {
			log.Warn().Err(err).Msg("group load from database failed")
			if err = loadFromEnvironment(g); err != nil {
				log.Warn().Err(err).Msg("group load from environment failed")
			} else {
				log.Info().Msg("created group")
			}
		}
	}

	var names []string
	for _, user := range g.Users {
		names = append(names, user.Name)
	}
	csv := strings.Join(names, ", ")
	log.Info().Msgf("created account group [%v]", csv)

	return g, nil
}

func loadFromJson(g *Group) error {
	if file, err := os.Open("pkg/config/users.json"); err != nil {
		// no account json file, not worth the log space
		return err
	} else if err := json.NewDecoder(file).Decode(&g); err != nil {
		log.Warn().Err(err).Msg("unable to decode users.json")
		return err
	} else if err := g.validate(); err != nil {
		log.Warn().Err(err).Msg("account json was invalid")
		return err
	} else {
		log.Info().Msg("created account group from account json")
		return nil
	}
}

func loadFromDatabase(g *Group) error {

	pg, err := db.OpenDB()
	if err != nil {
		return err
	}

	pg.Find(&g.Users)
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
		log.Error().Err(err).Msg("account environment variables were invalid")
		return err
	}
	log.Info().Msg("created account group from environment variables")
	return nil
}
