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
	"fmt"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"sort"
	"time"
)

func NewResult(session *config.Session, simulations []simulation, start time.Time) {

	log.Info().Msg(util.Tuna + " . ")
	log.Info().Msg(util.Tuna + " .. ")

	// sort by most successful net gain in asc order
	// so the best result is closest to the summary
	sort.SliceStable(simulations, func(i, j int) bool {
		return simulations[i].Net() < simulations[j].Net()
	})

	var trading, winners, losers, even int
	var sum, won, lost, net, volume float64
	for _, simulation := range simulations {

		if simulation.TotalEntries() == 0 {
			continue
		}

		productID := simulation.productID
		size := session.GetPattern(productID).Size

		log.Info().
			Float64(util.Delta, session.GetPattern(productID).Delta).
			Float64(util.Goal, session.GetPattern(productID).Gain).
			Float64(util.Quantity, size).
			Str(util.Link, util.CbUrl(productID)).
			Msg(util.Tuna + util.Break + util.GetCurrency(productID))

		log.Info().
			Str(util.Sigma, util.Usd(simulation.TotalAfterFees()*size)).
			Str(util.Quantity, util.Usd(simulation.TotalEntries()*size)).
			Str("%", util.Money(simulation.Net()*size)).
			Msg(util.Tuna + util.Break + fmt.Sprintf("%4s", simulation.symbol()))

		if simulation.WonLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.WonLen()).
				Str(util.Sigma, util.Usd(simulation.TotalWonAfterFees())).
				Str(util.Link, resultUrl(simulation.productID, "won", port())).
				Msg(util.Tuna + " ... " + fmt.Sprintf("%4s", util.Ice))
		}

		if simulation.LostLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.LostLen()).
				Str(util.Sigma, util.Usd(simulation.TotalLostAfterFees())).
				Str(util.Link, resultUrl(productID, "lst", port())).
				Msg(util.Tuna + " ... " + fmt.Sprintf("%5s", util.Poo))
		}

		if simulation.EvenLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.EvenLen()).
				Str(util.Sigma, "$0.000").
				Str(util.Link, resultUrl(productID, "evn", port())).
				Msg(util.Tuna + " ... " + fmt.Sprintf("%4s", util.Evn))
		}

		if simulation.TradingLen() > 0 {
			sum += simulation.TotalTradingAfterFees()
			symbol := util.TradingUp
			if simulation.TotalTradingAfterFees() < 0 {
				symbol = util.TradingDown
			}
			log.Info().
				Int(util.Quantity, simulation.TradingLen()).
				Str(util.Sigma, util.Usd(simulation.TotalTradingAfterFees())).
				Str(util.Link, resultUrl(productID, "dnf", port())).
				Msg(util.Tuna + " ... " + fmt.Sprintf("%4s", symbol))
		}

		winners += simulation.WonLen()
		losers += simulation.LostLen()
		trading += simulation.TradingLen()
		won += simulation.TotalWonAfterFees()
		lost += simulation.TotalLostAfterFees()
		net += simulation.TotalAfterFees() * session.GetPattern(simulation.productID).Size
		volume += simulation.TotalEntries() * session.GetPattern(simulation.productID).Size
		even += simulation.EvenLen()

		log.Info().Msg(util.Tuna + " ..")
	}

	if sum > 0 {
		won += sum
	} else {
		lost -= sum
	}

	log.Info().Msg(util.Tuna + " .")
	log.Info().Msg(util.Tuna + " ..")
	log.Info().Int("     "+util.Trading, trading).Msg(util.Tuna + " ...")
	log.Info().Int("     "+util.Lost, losers).Msg(util.Tuna + " ...")
	log.Info().Int("     "+util.Evn, even).Msg(util.Tuna + " ...")
	log.Info().Int("     "+util.Won, winners).Msg(util.Tuna + " ...")
	log.Info().Msg(util.Tuna + " ..")
	log.Info().Str("     "+util.Ice, util.Usd(won)).Msg(util.Tuna + " ...")
	log.Info().Str("     "+util.Poo, util.Usd(lost)).Msg(util.Tuna + " ...")
	log.Info().Str("     "+util.Net, util.Usd(net)).Msg(util.Tuna + " ...")
	log.Info().Msg(util.Tuna + " ..")
	log.Info().Str("     "+util.Volume, util.Usd(volume)).Msg(util.Tuna + " ...")
	log.Info().Str("      %", util.Money((net/volume)*100)).Msg(util.Tuna + " ...")
	log.Info().Msg(util.Tuna + " ..")
	log.Info().Msg(util.Tuna + " .")
	log.Info().Msg(util.Tuna + " ..")
	log.Info().Msgf("%s ... simulation generated in %f seconds", util.Tuna, time.Now().Sub(start).Seconds())
	log.Info().Msg(util.Tuna + " ..")
	log.Info().Msg(util.Tuna + " .")
	log.Info().Msg(util.Tuna + " .. ")
	log.Info().Msgf("%s ... charts home page http://localhost:%d", util.Tuna, port())
	log.Info().Msg(util.Tuna + " .. ")
	log.Info().Msg(util.Tuna + " . ")

}

func resultUrl(productID, dir string, port int) string {
	return fmt.Sprintf("http://localhost:%d/%s/%s.html", port, productID, dir)
}
