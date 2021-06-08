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
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"sort"
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

		positions, err := cbp.GetActivePositions()
		if err != nil {
			return err
		}

		var productIDs []string
		var cash, coin float64
		for productID, position := range positions {
			if position.Currency == "USD" {
				cash += position.Balance()
				continue
			}
			coin += position.Value()
			productIDs = append(productIDs, productID)
		}

		sort.Strings(productIDs)

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

		for _, productID := range productIDs {

			position := positions[productID]

			log.Info().
				Str(util.Sigma, util.Usd(position.Value())).
				Float64(util.Quantity, position.Balance()).
				Str(util.Hyperlink, position.Url()).
				Msg(util.Report + util.Break + position.Currency)

			orders, err := cbp.GetOrders(productID)
			if err != nil {
				return err
			}

			if len(*orders) > 0 {

				fills, err := cbp.GetFills(productID)
				if err != nil {
					return err
				}

				log.Info().Msg(util.Report + " ... held")

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

						t := order.CreatedAt.Time()
						from := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
						to := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 0, t.Location())

						rates, err := cbp.GetHistoricRates(productID, from, to)
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
						Str("3.", util.Usd(session.GetPattern(productID).GoalPrice(trade.Price()))).
						Str(util.Quantity, trade.Fill.Size).
						Msg(util.Report + " ... ")
				}
			}
			log.Info().Msg(util.Report + " ..")
		}

		log.Info().Msg(util.Report + " .")
		time.Sleep(time.Second * 30)
	}
}
