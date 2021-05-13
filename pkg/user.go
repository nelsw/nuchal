package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"nchl/config"
	"time"
)

type Portfolio struct {
	Username, Value string
	Positions       []Position
}

type Position struct {
	ProductId, Value string
	Balance          float64
}

func DisplayAccountInfo(cfg *config.Config) {

	dur, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		log.Error().Err(err)
		return
	}

	end := time.Now().Add(dur)
	for {
		displayAccountInfo(cfg.Users)
		time.Sleep(time.Minute)
		if time.Now().After(end) {
			fmt.Println("breaking")
			break
		}
	}
}

func displayAccountInfo(users []config.User) {
	for _, user := range users {

		var positions []Position
		var total float64
		for _, account := range GetAccounts(user) {

			if Float64(account.Balance) == 0.0 && Float64(account.Hold) == 0.0 {
				continue
			}

			productId := account.Currency + "-USD"
			balance := Float64(account.Balance)

			if account.Currency == "USD" {
				total += balance
				positions = append(positions, Position{
					ProductId: productId,
					Balance:   balance,
					Value:     Usd(balance),
				})
				continue
			}

			price := Float64(GetTicker(user, productId))
			value := price * balance
			total += value
			positions = append(positions, Position{
				ProductId: productId,
				Balance:   balance,
				Value:     Usd(value),
			})
		}

		portfolio := Portfolio{user.Name, Usd(total), positions}
		b, _ := json.MarshalIndent(&portfolio, "", "  ")
		fmt.Println(string(b))
	}
}
