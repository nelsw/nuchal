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
	SetUser(os.Args[3])
}

func SetUser(name string) {
	db.Where("name = ?", name).First(&user)
	if user == (User{}) {
		db.Where("name LIKE ?", "%"+name+"%").First(&user)
		if user == (User{}) {
			panic(errors.New(fmt.Sprintf("no user found for [%s]", name)))
		}
	}
}

func CreateUser() {
	u := User{
		Name:       os.Args[2],
		Key:        os.Args[3],
		Passphrase: os.Args[4],
		Secret:     os.Args[5],
	}
	fmt.Println("saving user", u.Name)
	db.Save(&u)
	fmt.Println("saved user", u.Name)
}
