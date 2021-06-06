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
	"github.com/rs/zerolog/log"
)

func NewDrops(session *config.Session) error {

	log.Info().Msg(util.Trade + " .")
	log.Info().Msg(util.Trade + " ..")
	log.Info().Msg(util.Trade + " ... trade --drop")
	log.Info().Msg(util.Trade + " ..")

	for _, productId := range *session.ProductIds() {

		log.Info().Msg(util.Trade + " ... " + productId)

		orders, err := session.GetOrders(productId)
		if err != nil {
			return err
		}

		for _, order := range *orders {
			if err := session.CancelOrder(order.ID); err != nil {
				return err
			}
			log.Info().Msg(util.Trade + " ... dropped")
		}
		log.Info().Msg(util.Trade + " ..")
	}
	log.Info().Msg(util.Trade + " .")

	return nil

}
