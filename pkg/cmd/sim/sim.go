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

package sim

import (
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"time"
)

// New creates a new simulation, and boy is that an understatement.
// Per usual, we start by getting program configurations.
func New(session *config.Session, winnersOnly, noLosers bool) error {

	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " ... simulation")
	log.Info().Msg(util.Sim + " ..")

	var simulations []simulation

	start := time.Now()
	alpha := *session.Alpha
	omega := *session.Omega
	pg := db.NewDB(&cbp.Rate{})

	for _, productID := range session.UsdSelectionProductIDs() {

		var rates []cbp.Rate
		pg.Where("product_id = ?", productID).
			Where("unix BETWEEN ? AND ?", alpha.UnixNano(), omega.UnixNano()).
			Order("unix asc").
			Find(&rates)

		if len(rates) == 0 ||
			rates[0].Time().Sub(alpha).Minutes() > 3 ||
			rates[len(rates)-1].Time().Sub(omega).Minutes() > 3 {

			if out, err := cbp.Rates(productID, session.RateParams()); err != nil {
				return err
			} else {
				log.Debug().
					Int("coinbase", len(out)).
					Msgf("%s ... %s ... %s", util.Sim, util.GetCurrency(productID), util.Check)
				for _, rate := range out {
					rates = append(rates, rate)
					pg.Create(&rate)
				}
			}
		} else {
			log.Debug().
				Int("database", len(rates)).
				Msgf("%s ... %s ... %s", util.Sim, util.GetCurrency(productID), util.Check)
		}

		var s simulation
		newSimulation(session, productID, rates, &s)
		if s.TotalEntries() == 0 ||
			((noLosers || winnersOnly) && s.LostLen() > 0) ||
			winnersOnly && s.TradingLen() > 0 {
			continue
		}

		simulations = append(simulations, s)

		log.Info().Msg(util.Sim + util.Break + util.GetCurrency(productID) + util.Break + util.ChequeredFlag)
		log.Info().Msg(util.Sim + " ..")
	}
	log.Info().Msg(util.Sim + " ..")

	go NewResult(session, simulations, start)

	return newSite(simulations)
}
