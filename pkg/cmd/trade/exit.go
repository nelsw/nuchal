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
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
)

// NewExits sells every active trading position.
func NewExits(session *config.Session) error {

	log.Info().Msg(util.Shark + " .")
	log.Info().Msg(util.Shark + " ..")
	log.Info().Msg(util.Shark + " ... trade --exit")
	log.Info().Msg(util.Shark + " ..")
	log.Info().Msg(util.Shark + " .")

	positions, err := cbp.GetTradingPositions()
	if err != nil {
		return err
	}

	if len(positions) < 1 {
		log.Info().Msg(util.Shark + " ..")
		log.Info().Msg(util.Shark + " ...")
		log.Info().Msg(util.Shark + " ... no available balance found to exit")
		log.Info().Msg(util.Shark + " ...")
		log.Info().Msg(util.Shark + " ..")
		log.Info().Msg(util.Shark + " .")
		return nil
	}

	for _, productID := range session.UsdSelectionProductIDs() {

		currency := util.GetCurrency(productID)

		log.Info().Msg(util.Shark + " ... " + currency + util.Break + "exit")

		position := positions[productID]

		for _, trade := range position.GetActiveTrades() {

			order := session.GetPattern(productID).NewMarketSellOrder(trade.Fill.Size)
			if _, err := cbp.CreateOrder(order); err != nil {
				return err
			}
			log.Info().Msg(util.Shark + " ... " + currency + util.Break + "exited")
		}
		log.Info().Msg(util.Shark + " ..")
	}
	log.Info().Msg(util.Shark + " .")

	return nil
}
