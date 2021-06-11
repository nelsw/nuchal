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
	"math"
	"sort"
	"time"
)

func NewResult(session *config.Session, results []simulation, start time.Time) {

	log.Info().Msg(util.Sim + " . ")
	log.Info().Msg(util.Sim + " .. ")

	// sort by most successful net gain in asc order
	// so the best result is closest to the summary
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Net() < results[j].Net()
	})

	var trading, winners, losers, even int
	var sum, won, lost, net, volume float64
	for _, simulation := range results {

		if simulation.TotalEntries() == 0 {
			continue
		}

		productID := simulation.productID
		size := session.GetPattern(productID).Size

		log.Info().
			Float64(util.Delta, session.GetPattern(productID).Delta).
			Float64(util.UpArrow, session.GetPattern(productID).Gain).
			Float64(util.Quantity, size).
			Str(util.Hyperlink, util.CbUrl(productID)).
			Msg(util.Sim + util.Break + util.GetCurrency(productID))

		log.Info().
			Str(util.Sigma, util.Usd(simulation.TotalAfterFees()*size)).
			Str(util.Quantity, util.Usd(simulation.TotalEntries()*size)).
			Str("%", util.Money(simulation.Net()*size)).
			Msg(util.Sim + util.Break + fmt.Sprintf("%4s", simulation.symbol()))

		if simulation.WonLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.WonLen()).
				Str(util.Sigma, util.Usd(simulation.TotalWonAfterFees())).
				Str(util.Hyperlink, resultUrl(simulation.productID, "won", port())).
				Msg(util.Sim + " ... " + fmt.Sprintf("%4s", util.Diamond))
		}

		if simulation.LostLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.LostLen()).
				Str(util.Sigma, util.Usd(simulation.TotalLostAfterFees())).
				Str(util.Hyperlink, resultUrl(productID, "lst", port())).
				Msg(util.Sim + " ... " + fmt.Sprintf("%5s", util.Lost))
		}

		if simulation.EvenLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.EvenLen()).
				Str(util.Sigma, "$0.000").
				Str(util.Hyperlink, resultUrl(productID, "evn", port())).
				Msg(util.Sim + " ... " + fmt.Sprintf("%4s", util.Even))
		}

		if simulation.TradingLen() > 0 {
			sum += simulation.TotalTradingAfterFees()
			symbol := util.UpTrend
			if simulation.TotalTradingAfterFees() < 0 {
				symbol = util.DnTrend
			}
			log.Info().
				Int(util.Quantity, simulation.TradingLen()).
				Str(util.Sigma, util.Usd(simulation.TotalTradingAfterFees())).
				Str(util.Hyperlink, resultUrl(productID, "dnf", port())).
				Msg(util.Sim + " ... " + fmt.Sprintf("%4s", symbol))
		}

		winners += simulation.WonLen()
		losers += simulation.LostLen()
		trading += simulation.TradingLen()
		won += simulation.TotalWonAfterFees()
		lost += simulation.TotalLostAfterFees()
		net += simulation.TotalAfterFees() * session.GetPattern(simulation.productID).Size
		volume += simulation.TotalEntries() * session.GetPattern(simulation.productID).Size
		even += simulation.EvenLen()

		log.Info().Msg(util.Sim + " ..")
	}

	if sum > 0 {
		won += sum
	} else {
		lost -= sum
	}

	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Int("    trading", trading).Msg(util.Sim + " ...")
	log.Info().Int("       lost", losers).Msg(util.Sim + " ...")
	log.Info().Int("       even", even).Msg(util.Sim + " ...")
	log.Info().Int("        won", winners).Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Str("       lost", util.Usd(lost)).Msg(util.Sim + " ...")
	log.Info().Str("        won", util.Usd(won)).Msg(util.Sim + " ...")
	log.Info().Str("        net", util.Usd(net)).Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Str("     volume", util.Usd(volume)).Msg(util.Sim + " ...")
	log.Info().Str("          %", util.Money((math.Min(1, net)/math.Min(1, volume))*100)).Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msgf("%s ... simulation generated in %f seconds", util.Sim, time.Now().Sub(start).Seconds())
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " .. ")
	log.Info().Msgf("%s ... charts home page http://localhost:%d", util.Sim, port())
	log.Info().Msg(util.Sim + " .. ")
	log.Info().Msg(util.Sim + " . ")

}

func resultUrl(productID, dir string, port int) string {
	return fmt.Sprintf("http://localhost:%d/%s/%s.html", port, productID, dir)
}
