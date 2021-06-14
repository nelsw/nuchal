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

		log.Info().Msg(util.Puffer + " .")
		log.Info().Msg(util.Puffer + " ..")
		log.Info().Msg(util.Puffer + " ...")
		log.Info().Msg(util.Puffer + " ... report")
		log.Info().Msg(util.Puffer + " ...")
		log.Info().Msg(util.Puffer + " ..")
		log.Info().Msg(util.Puffer + " .")

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

		log.Info().Msg(util.Puffer + " ..")
		log.Info().Msg(util.Puffer + " ... portfolio")
		log.Info().Str(util.Dollar, dollar).Str(util.Currency, currency).Str(util.Sigma, sigma).Msg(util.Puffer + " ...")
		log.Info().Msg(util.Puffer + " ..")
		log.Info().Msg(util.Puffer + " .")
		log.Info().Msg(util.Puffer + " ..")
		log.Info().Msg(util.Puffer + " ... positions")
		log.Info().Msg(util.Puffer + " ..")

		for _, productID := range productIDs {

			position := positions[productID]

			log.Info().
				Str(util.Sigma, util.Usd(position.Value())).
				Float64(util.Quantity, position.Balance()).
				Str(util.Link, util.CbUrl(productID)).
				Msg(util.Puffer + util.Break + util.GetCurrency(productID))

			orders, err := cbp.GetOrders(productID)
			if err != nil {
				return err
			}

			pattern := session.GetPattern(productID)

			if len(*orders) > 0 {

				fills, err := cbp.GetFills(productID)
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
						Str(util.Entry, pattern.PrecisePrice(entryPrice)).
						Str(util.Current, pattern.PrecisePrice(position.Price())).
						Str(util.Goal, pattern.PrecisePriceFromString(order.Price)).
						Str(util.Quantity, pattern.PreciseSize(order.Size)).
						Time(util.Time, order.CreatedAt.Time()).
						Msg(util.Puffer + util.Break + "   " + util.Hold)
				}
			}

			trades := position.GetActiveTrades()
			if len(trades) > 0 {
				for _, trade := range trades {
					log.Info().
						Str(util.Entry, pattern.PrecisePrice(trade.Price())).
						Str(util.Current, pattern.PrecisePrice(position.Price())).
						Str(util.Goal, pattern.PrecisePrice(pattern.GoalPrice(trade.Price()))).
						Str(util.Quantity, pattern.PreciseSize(trade.Fill.Size)).
						Time(util.Time, trade.CreatedAt.Time()).
						Msg(util.Puffer + util.Break + "   " + util.Trading)
				}
			}
			log.Info().Msg(util.Puffer + " ..")
		}

		log.Info().Msg(util.Puffer + " .")
		time.Sleep(time.Second * 30)
	}
}
