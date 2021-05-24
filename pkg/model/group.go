package model

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"nuchal/pkg/db"
	"os"
	"strings"
)

type Group struct {
	Users []User `json:"users"`
}

func (c Group) GetUser(name string) (*User, error) {
	for _, user := range c.Users {
		if user.Name == name {
			return &user, nil
		}
	}
	return nil, errors.New("no user found for " + name)
}

func NewGroup() (*Group, error) {

	log.Info().Msg("configuring user group")

	g := new(Group)

	if file, err := os.Open("pkg/config/users.json"); err == nil {
		if err := json.NewDecoder(file).Decode(&g); err != nil {
			return nil, err
		}
	} else {
		db.NewDB().Where("enable = ?", true).Find(&g.Users)
	}

	if g.Users == nil && len(g.Users) < 1 {
		usr, err := NewUser()
		if err != nil {
			return nil, err
		}
		g.Users = append(g.Users, *usr)
	}

	var names []string
	var enabledUsers []User
	for _, user := range g.Users {
		if !user.Enable {
			continue
		}
		if err := user.validate(); err != nil {
			return nil, err
		}
		enabledUsers = append(enabledUsers, user)
		names = append(names, user.Name)
	}
	g.Users = enabledUsers

	csv := strings.Join(names, ", ")
	log.Info().Msgf("configured user group [%v]", csv)

	return g, nil
}
