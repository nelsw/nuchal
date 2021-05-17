package status

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	config2 "nchl/config"
	"nchl/pkg/model/account"
	"nchl/pkg/util"
	"sort"
	"time"
)

func New() error {

	log.Info().Msg("get status")

	cfg, err := config2.NewConfig()
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		return err
	}

	for {

		for _, user := range cfg.Users {

			portfolio, err := getPortfolio(user)
			if err != nil { // logged in getPortfolio
				return err
			}

			log.Info().
				Str("username", portfolio.Username).
				Str("value", portfolio.Value).
				Msg("portfolio")

			for _, position := range portfolio.Positions {

				log.Info().
					Str("productId", position.ProductId).
					Str("value", util.Usd(position.Value)).
					Msg("position")

				if position.Fills != nil && len(position.Fills) > 0 {

					log.Warn().
						Float64("balance", position.Balance).
						Float64("hold", position.Hold).
						Msg("position")

					posture := cfg.GetPosture(position.ProductId)
					for _, fill := range position.Fills {

						size := fill.Size
						price := util.Float64(fill.Price)
						price += price * posture.GainFloat()

						if _, err := user.GetClient().CreateOrder(posture.StopEntryOrder(price, size)); err != nil {
							log.Error().Err(err).Msg("creating order")
						}
					}
				}
			}
		}

		log.Info().Msg("got status")

		if util.IsTestMode() || cfg.IsTimeToExit() {
			return nil
		}

		time.Sleep(time.Minute)
	}
}

func getPortfolio(u account.User) (*account.Portfolio, error) {

	accounts, err := u.GetClient().GetAccounts()
	if err != nil { // logged in user
		return nil, err
	}

	var positions []account.Position

	for _, a := range accounts {

		if hasNoPosition(a) {
			continue
		}

		position, err := getPosition(u, a)
		if err != nil { // logged get Position
			return nil, err
		}

		positions = append(positions, *position)
	}

	return account.NewPortfolio(u.Name, positions), nil
}

func getPosition(u account.User, a cb.Account) (*account.Position, error) {

	balance := util.Float64(a.Balance)

	if a.Currency == "USD" {
		return &account.Position{
			ProductId: "USD",
			Balance:   balance,
			Hold:      0.0,
			Value:     balance,
		}, nil
	}

	productId := a.Currency + "-USD"

	var value float64
	if ticker, err := u.GetClient().GetTicker(productId); err != nil {
		return nil, err
	} else {
		value = util.Float64(ticker.Price) * balance
	}

	hold := util.Float64(a.Hold)
	if balance == hold {
		return &account.Position{
			ProductId: productId,
			Balance:   balance,
			Hold:      balance,
			Value:     value,
		}, nil
	}

	recentBuyFills, err := getRecentBuyFills(u, productId)
	if err != nil {
		return nil, err
	}

	var fills []cb.Fill
	for _, fill := range recentBuyFills {
		fills = append(fills, fill)
		hold += util.Float64(fill.Size)
		if balance == hold {
			break
		}
	}

	return &account.Position{productId, value, balance, util.Float64(a.Hold), fills}, nil
}

func hasNoPosition(a cb.Account) bool {
	return util.IsZero(a.Balance) && util.IsZero(a.Hold)
}

func getRecentBuyFills(u account.User, productId string) ([]cb.Fill, error) {

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

	sort.SliceStable(allFills, func(i, j int) bool {
		return allFills[i].CreatedAt.Time().Before(allFills[j].CreatedAt.Time())
	})

	var buys, sells []cb.Fill

	for _, fill := range allFills {
		if fill.Side == "buy" {
			buys = append(buys, fill)
		} else {
			sells = append(sells, fill)
		}
	}

	qty := util.MinInt(len(buys), len(sells))
	result := buys[qty:]

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.Time().After(result[j].CreatedAt.Time())
	})

	return result, nil
}
