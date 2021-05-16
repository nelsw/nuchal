package status

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"nchl/pkg/account"
	"nchl/pkg/nuchal"
	"nchl/pkg/util"
	"time"
)

func New() error {
	log.Info().Msg("creating status")
	if c, err := nuchal.NewConfig(); err != nil {
		log.Error().Err(err)
		return err
	} else {
		for {
			displayAccountInfo(c.Group)
			if c.TestMode || c.IsTimeToExit() {
				log.Info().Msg("created status")
				return nil
			}
			time.Sleep(time.Minute)
		}
	}
}

func displayAccountInfo(g *account.Group) {

	for _, u := range g.Users {

		accounts, err := u.GetClient().GetAccounts()
		if err != nil {
			log.Error().Err(err).Msg("error getting account for " + u.Name)
			continue
		}

		var positions []account.Status
		var total float64
		for _, a := range accounts {

			if util.Float64(a.Balance) == 0.0 && util.Float64(a.Hold) == 0.0 {
				continue
			}

			productId := a.Currency + "-USD"
			balance := util.Float64(a.Balance)

			if a.Currency == "USD" {
				total += balance
				positions = append(positions, account.Status{
					ProductId: productId,
					Balance:   balance,
					Value:     util.Usd(balance),
				})
				continue
			}

			ticker, err := u.GetClient().GetTicker(productId)
			if err != nil {
				log.Warn().Err(err).Msg("error getting ticker price for " + productId)
				continue
			}

			value := util.Float64(ticker.Price) * balance
			total += value
			positions = append(positions, account.Status{
				ProductId: productId,
				Balance:   balance,
				Value:     util.Usd(value),
			})
		}

		portfolio := account.Portfolio{u.Name, util.Usd(total), positions}
		b, _ := json.MarshalIndent(&portfolio, "", "  ")
		fmt.Println(string(b))
	}
}
