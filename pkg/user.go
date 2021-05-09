package pkg

import (
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
	"os"
)

var user User

type User struct {
	gorm.Model
	Name       string `json:"name"`
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Secret     string `json:"secret"`
}

func init() {
	if err := db.AutoMigrate(user); err != nil {
		panic(err)
	}
}

func SetupUser() {
	name := os.Args[3]
	db.Where("name = ?", name).First(&user)
	if user == (User{}) {
		panic(errors.New(fmt.Sprintf("no user found for [%s]", name)))
	}
}