package pkg

import (
	"fmt"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
	"nchl/pkg/db"
)

type User struct {
	gorm.Model
	Name       string `json:"name"`
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
	Secret     string `json:"secret"`
}

const (
	eqQuery = "name = ?"
	lkQuery = "name LIKE ?"
)

func init() {
	var user User
	if err := db.Client.AutoMigrate(user); err != nil {
		panic(err)
	}
}

func GetUserConfig(username string) (*string, *string, *string) {
	var user User
	db.Client.Where(eqQuery, username).First(&user)
	if user == (User{}) {
		db.Client.Where(lkQuery, "%"+username+"%").First(&user)
		if user == (User{}) {
			panic(fmt.Sprintf("no user found where name = [%s]", username))
		}
	}
	return &user.Key, &user.Passphrase, &user.Secret
}

func CreateUser(username, key, pass, secret string) {
	fmt.Println("creating user")
	db.Client.Save(User{
		Name:       username,
		Key:        key,
		Passphrase: pass,
		Secret:     secret,
	})
	fmt.Println("created user")
}
