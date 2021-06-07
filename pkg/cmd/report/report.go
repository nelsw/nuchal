/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package report

import (
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"time"
)

// New creates a new report.
func New(session *config.Session) error {

	for {

		log.Info().Msg(util.Report + " .")
		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " ...")
		log.Info().Msg(util.Report + " ... report")
		log.Info().Msg(util.Report + " ...")
		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " .")

		positions, err := session.GetActivePositions()
		if err != nil {
			return err
		}

		var cash, coin float64
		for _, position := range *positions {
			if position.Currency == "USD" {
				cash += position.Balance()
				continue
			}
			coin += position.Value()
		}

		dollar := util.Money(cash)
		currency := util.Money(coin)
		sigma := util.Usd(cash + coin)

		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " ... portfolio")
		log.Info().Str(util.Dollar, dollar).Str(util.Currency, currency).Str(util.Sigma, sigma).Msg(util.Report + " ...")
		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " .")
		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " ... positions")
		log.Info().Msg(util.Report + " ..")

		for _, position := range *positions {

			if position.Currency == "USD" {
				continue
			}

			log.Info().
				Str(util.Sigma, util.Usd(position.Value())).
				Float64(util.Quantity, position.Balance()).
				Str(util.Hyperlink, position.Url()).
				Msg(util.Report + util.Break + position.Currency)

			orders, err := session.GetOrders(position.ProductId())
			if err != nil {
				return err
			}

			if len(*orders) > 0 {
				log.Info().Msg(util.Report + " ... held")
				fills, err := session.GetFills(position.ProductId())
				if err != nil {
					return err
				}

				for orderIdx, order := range *orders {

					if order.Side == "buy" {
						continue
					}

					var entryPrice float64
					for fillIdx, fill := range *fills {

						if fill.Side != "buy" {
							continue
						}

						if orderIdx == fillIdx { // direct match?
							entryPrice = util.Float64(fill.Price)
							break
						}

						productID := order.ProductID

						t := order.CreatedAt.Time()
						from := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
						to := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 0, t.Location())

						params := cb.GetHistoricRatesParams{from, to, 60}

						rates, err := session.GetClient().GetHistoricRates(productID, params)
						if err != nil {
							return err
						}
						if rates == nil || len(rates) < 1 {
							entryPrice = -1
						} else {
							entryPrice = rates[0].Open
						}

						break
					}

					log.Info().
						Time("", order.CreatedAt.Time()).
						Str("1.", util.Usd(entryPrice)).
						Str("2.", util.Usd(position.Price())).
						Str("3.", util.Usd(util.Float64(order.Price))).
						Str(util.Quantity, order.Size).
						Msg(util.Report + " ... ")
				}
			}

			trades := position.GetActiveTrades()
			if len(trades) > 0 {
				log.Info().Msg(util.Report + " ... active")
				for _, trade := range trades {
					log.Info().
						Time("", trade.CreatedAt.Time()).
						Str("1.", util.Usd(trade.Price())).
						Str("2.", util.Usd(position.Price())).
						Str("3.", util.Usd(position.GoalPrice(trade.Price()))).
						Float64(util.Quantity, trade.Size()).
						Msg(util.Report + " ... ")
				}
			}
			log.Info().Msg(util.Report + " ..")
		}

		log.Info().Msg(util.Report + " .")
		time.Sleep(time.Second * 30)
	}
}
