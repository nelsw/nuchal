package model

import (
	"encoding/json"
	"errors"
	"github.com/nelsw/nuchal/pkg/db"
	"os"
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

	var enabledUsers []User
	for _, user := range g.Users {
		if !user.Enable {
			continue
		}
		if err := user.validate(); err != nil {
			return nil, err
		}
		enabledUsers = append(enabledUsers, user)
	}
	g.Users = enabledUsers

	return g, nil
}
