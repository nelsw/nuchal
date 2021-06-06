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
	"fmt"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"time"
)

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

			dollar = util.Money(position.Price())
			sigma = util.Usd(position.Value())
			qty := position.Balance()
			msg := util.Report + " ... " + position.Currency

			log.Info().Str(util.Dollar, dollar).Str(util.Sigma, sigma).Float64(util.Balance, qty).Msg(msg)

			orders, err := session.GetOrders(position.ProductId())
			if err != nil {
				return err
			}

			if len(*orders) > 0 {

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

						productId := order.ProductID

						t := order.CreatedAt.Time()
						from := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
						to := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 0, t.Location())

						params := cb.GetHistoricRatesParams{from, to, 60}

						rates, err := session.GetClient().GetHistoricRates(productId, params)
						if err != nil {
							return err
						}
						entryPrice = rates[0].Open
						break
					}

					log.Info().
						Str(util.Dollar, util.Money(entryPrice)).
						Str(util.Target, util.Money(util.Float64(order.Price))).
						Str(util.Balance, fmt.Sprintf("%.0f", util.Float64(order.Size))).
						Time(util.Time, order.CreatedAt.Time()).
						Msg(util.Report + " ... hold")
				}
			}

			trades := position.GetActiveTrades()
			if len(trades) > 0 {
				log.Info().Msg(util.Report + " ... active")
				for _, trade := range trades {

					goal := position.GoalPrice(trade.Price())
					net := (goal - (goal * session.Maker)) * trade.Size()

					log.Info().
						Str(util.Dollar, fmt.Sprintf("%.3f", trade.Price())).
						Str(util.Balance, fmt.Sprintf("%.0f", trade.Size())).
						Time(util.Time, trade.CreatedAt.Time()).
						Str(util.Target, fmt.Sprintf("%.3f", goal)).
						Str(util.Profit, fmt.Sprintf("%.3f", net)).
						Msg(util.Report + " ... ")
				}
			}
			log.Info().Msg(util.Report + " ..")
		}

		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " .")

		time.Sleep(time.Second * 30)
	}
}
