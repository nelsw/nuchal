package account

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal/pkg/config"
	"nuchal/pkg/model/account"
	"nuchal/pkg/util"
	"time"
)

func New() error {
	return NewWithForceHolds(false)
}

func NewWithForceHolds(forceHolds bool) error {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	for {

		for _, user := range cfg.Users {

			log.Info().Msg("---------")

			portfolio, err := getPortfolio(user)
			if err != nil {
				return err
			}

			portfolio.Info()

			for _, position := range portfolio.CoinPositions() {

				product := cfg.GetProduct(position.ProductId)
				position.Log(product)

				if position.HasOrphanBuyFills() && forceHolds {

					posture := cfg.GetPosture(product.ID)

					for _, fill := range position.OrphanBuyFills() {

						order := posture.StopGainOrder(fill)

						if _, err := user.GetClient().CreateOrder(order); err != nil {
							return err
						}
					}
				}
			}
		}

		if util.IsTestMode() || cfg.IsTimeToExit() {
			return nil
		}

		util.Sleep(time.Minute * 3)
	}
}

func getPortfolio(u account.User) (*account.Portfolio, error) {

	accounts, err := u.GetClient().GetAccounts()
	if err != nil { // logged in account
		return nil, err
	}

	var positions []account.Position

	for _, a := range accounts {

		if util.IsZero(a.Balance) && util.IsZero(a.Hold) {
			continue
		}

		position, err := getPosition(u, a)
		if err != nil {
			return nil, err
		}

		positions = append(positions, *position)
	}

	return account.NewPortfolio(u.Name, positions), nil
}

func getPosition(u account.User, a cb.Account) (*account.Position, error) {

	if a.Currency == "USD" {
		return account.NewUsdPosition(a.Balance), nil
	}

	productId := a.Currency + "-USD"
	cursor := u.GetClient().ListFills(cb.ListFillsParams{ProductID: productId})

	var newFills, allFills []cb.Fill
	for cursor.HasMore {

		if err := cursor.NextPage(&newFills); err != nil {
			return nil, err
		}

		for _, chunk := range newFills {
			allFills = append(allFills, chunk)
		}
	}

	balance := util.Float64(a.Balance)

	var value float64
	if ticker, err := u.GetClient().GetTicker(productId); err != nil {
		return nil, err
	} else {
		value = util.Float64(ticker.Price) * balance
	}

	return account.NewPosition(productId, a.Hold, value, balance, allFills), nil
}
