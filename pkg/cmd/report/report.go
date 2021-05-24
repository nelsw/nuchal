package report

import (
	"fmt"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/model"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"time"
)

func New(username string, forceHolds, recurring bool) error {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	util.PrintNewLine()

	if username != util.GuestName {
		if user, err := cfg.GetUser(username); err == nil {
			err = printPortfolio(cfg, *user, forceHolds)
			if err == nil {
				return nil
			}
		}
	}

	for {

		for _, user := range cfg.Users {
			if err := printPortfolio(cfg, user, forceHolds); err != nil {
				return err
			}
		}

		if !recurring {
			return nil
		}

		util.LogBanner()
		util.Sleep(time.Minute * 1)
	}
}

func printPortfolio(cfg *config.Config, user model.User, forceHolds bool) error {

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

				if o, err := user.GetClient().CreateOrder(order); err != nil {
					if util.IsInsufficientFunds(err) {
						continue
					}
					return err
				} else {
					fmt.Println(o)
				}
			}
		}
	}
	util.PrintNewLine()
	return nil
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
