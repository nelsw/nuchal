package report

import (
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"time"
)

func New() error {

	cfg, err := config.NewSession()
	if err != nil {
		return err
	}

	util.PrintNewLine()

	for {

		positions, err := cfg.GetActivePositions()
		if err != nil {
			return err
		}

		cash := 0.0
		crypto := 0.0
		for _, position := range *positions {
			if position.Currency == "USD" {
				cash += position.Balance()
				continue
			}
			crypto += position.Value()
		}

		log.Info().
			Str(util.Dollar, util.Usd(cash)).
			Str(util.Currency, util.Usd(crypto)).
			Str(util.Sigma, util.Usd(cash+crypto)).
			Msg(cfg.User())

		for _, position := range *positions {
			if position.Currency != "USD" {
				position.Log()
			}
		}

		util.PrintNewLine()
		util.LogBanner()

		if cfg.End().After(time.Now()) {
			return nil
		}

		util.Sleep(time.Minute * 1)
	}
}
