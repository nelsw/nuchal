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

package trade

import (
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
)

// NewHolds will create a limit sell entry order for every trade that does not already meet or exceed the goal price.
func NewHolds(session *config.Session) error {

	log.Info().Msg(util.Trade + " .")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " ... trade --hold")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " .")

	positions, err := session.GetTradingPositions()
	if err != nil {
		return err
	}

	if len(positions) < 1 {
		log.Info().Msg(util.Trade + " ..")
		log.Info().Msg(util.Trade + " ...")
		log.Info().Msg(util.Trade + " ... no available balance found to hold")
		log.Info().Msg(util.Trade + " ...")
		log.Info().Msg(util.Trade + " ..")
		log.Info().Msg(util.Trade + " .")
		return nil
	}

	for _, position := range positions {

		log.Info().Msg(util.Trade + " ... " + position.ProductId())
		for _, trade := range position.GetActiveTrades() {

			currentPrice, err := session.GetCurrentPrice(trade.ProductID)
			if err != nil {
				return err
			}

			var order *cb.Order
			goalPrice := position.GoalPrice(trade.Price())
			if *currentPrice >= goalPrice {
				order = position.NewMarketSellOrder(trade.Fill.Size)
			} else {
				order = position.NewLimitSellEntryOrder(goalPrice, trade.Fill.Size)
			}

			if _, err := session.CreateOrder(order); err != nil {
				return err
			}

			log.Info().Msg(util.Trade + " ... held")
		}
		log.Info().Msg(util.Trade + " ..")
	}
	log.Info().Msg(util.Trade + " .")

	return nil
}
