package report

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal/pkg/config"
	"nuchal/pkg/model"
	"nuchal/pkg/util"
	"time"
)

func New(forceHolds, recurring bool) error {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	util.PrintNewLine()

	for {

		for _, user := range cfg.Users {

			portfolio, err := getPortfolio(user)
			if err != nil {
				return err
			}

			portfolio.Info()

			for _, position := range portfolio.CoinPositions() {

				position.Log()

				if position.HasOrphanBuyFills() && forceHolds {

					posture := cfg.GetPosture(position.ProductId())

					for _, fill := range position.OrphanBuyFills() {

						order := posture.StopGainOrder(fill)

						if _, err := user.GetClient().CreateOrder(order); err != nil {
							return err
						}
					}
				}
			}
			util.PrintNewLine()
		}

		util.LogBanner()
		util.PrintCursor()

		if !recurring {
			return nil
		}

		util.Sleep(time.Minute * 1)
	}
}

func getPortfolio(u model.User) (*model.Portfolio, error) {

	accounts, err := u.GetClient().GetAccounts()
	if err != nil {
		return nil, err
	}

	var positions []model.Position

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

	return model.NewPortfolio(u.Name, positions), nil
}

func getPosition(u model.User, a cb.Account) (*model.Position, error) {

	if a.Currency == "USD" {
		return model.NewPosition(a, cb.Ticker{}, nil), nil
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

	ticker, err := u.GetClient().GetTicker(productId)
	if err != nil {
		return nil, err
	}

	return model.NewPosition(a, ticker, allFills), nil
}
