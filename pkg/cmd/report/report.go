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
	"github.com/rs/zerolog/log"
	"time"
)

func New(usd []string, size, gain, loss, delta float64) error {

	ses, err := config.NewSession(usd, size, gain, loss, delta)
	if err != nil {
		return err
	}

	for {

		log.Info().Msg(util.Report + " .")
		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " ...")
		log.Info().Msg(util.Report + " ... report")
		log.Info().Msg(util.Report + " ...")
		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " .")

		positions, err := ses.GetActivePositions()
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

		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " ... portfolio")
		log.Info().
			Str(util.Dollar, util.Money(cash)).
			Str(util.Currency, util.Money(coin)).
			Str(util.Sigma, util.Usd(cash+coin)).
			Msg(util.Report + " ...")
		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " ... positions")
		log.Info().Msg(util.Report + " ..")

		for _, position := range *positions {

			if position.Currency == "USD" {
				continue
			}

			log.Info().
				Str(util.Dollar, util.Money(position.Price())).
				Str(util.Sigma, util.Usd(position.Value())).
				Float64(util.Quantity, position.Balance()).
				Msg(util.Report + " ... " + position.Currency)

			out, err := ses.GetOrders(position.ProductId())
			if err != nil {
				return err
			}

			for _, order := range *out {
				log.Info().
					Str(util.Dollar, util.Money(util.Float64(order.Price))).
					Str(util.Quantity, fmt.Sprintf("%.0f", util.Float64(order.Size))).
					Time("🗓", order.CreatedAt.Time()).
					Msg(util.Report + " ... hold")
			}

			for _, trade := range position.GetActiveTrades() {

				goal := position.GoalPrice(trade.Price())
				net := (goal - (goal * ses.Maker)) * trade.Size()

				log.Warn().
					Str(util.Dollar, fmt.Sprintf("%.3f", trade.Price())).
					Str(util.Quantity, fmt.Sprintf("%.0f", trade.Size())).
					Time("🗓", trade.CreatedAt.Time()).
					Str("🎯", fmt.Sprintf("%.3f", goal)).
					Str("💰", fmt.Sprintf("%.3f", net)).
					Msg(util.Report + " ... sell")
			}

			log.Info().Msg(util.Report + " ..")
		}

		log.Info().Msg(util.Report + " ..")
		log.Info().Msg(util.Report + " .")

		time.Sleep(time.Second * 15)
	}
}
