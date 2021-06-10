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
	"github.com/nelsw/nuchal/pkg/cbp"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/db"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"sort"
	"time"
)

// New creates a new simulation, and boy is that an understatement.
// Per usual, we start by getting program configurations.
func New(session *config.Session, winnersOnly, noLosers bool) error {

	log.Info().Msg(util.Sim + " .")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " ... simulation")
	log.Info().Msg(util.Sim + " ..")

	var results []simulation
	for _, productID := range session.UsdSelectionProductIDs() {

		rates, err := getRates(session, productID)
		if err != nil {
			return err
		}

		simulation := newSimulation(session, productID, rates)
		log.Info().Msg(util.Sim + util.Break + util.GetCurrency(productID) + util.Break + util.ChequeredFlag)
		log.Info().Msg(util.Sim + " .. ")

		if simulation.TotalEntries() == 0 {
			continue
		}

		if (noLosers || winnersOnly) && simulation.LostLen() > 0 {
			continue
		}

		if winnersOnly && simulation.TradingLen() > 0 {
			continue
		}

		results = append(results, *simulation)
	}

	log.Info().Msg(util.Sim + " . ")
	log.Info().Msg(util.Sim + " .. ")

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Net() < results[j].Net()
	})

	var trading, winners, losers, even int
	var sum, won, lost, net, volume float64
	for _, simulation := range results {

		productID := simulation.productID
		size := session.GetPattern(productID).Size

		log.Info().
			Float64(util.Delta, session.GetPattern(productID).Delta).
			Float64(util.UpArrow, session.GetPattern(productID).Gain).
			Float64(util.Quantity, size).
			Str(util.Hyperlink, util.CbUrl(productID)).
			Msg(util.Sim + util.Break + util.GetCurrency(productID))

		if simulation.WonLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.WonLen()).
				Str(util.Sigma, util.Usd(simulation.TotalWonAfterFees())).
				Str(util.Hyperlink, resultUrl(simulation.productID, "won", port())).
				Msg(util.Sim + " ... " + fmt.Sprintf("%4s", util.ThumbsUp))
		}

		if simulation.LostLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.LostLen()).
				Str(util.Sigma, util.Usd(simulation.TotalLostAfterFees())).
				Str(util.Hyperlink, resultUrl(productID, "lst", port())).
				Msg(util.Sim + " ... " + fmt.Sprintf("%5s", util.ThumbsDn))
		}

		if simulation.EvenLen() > 0 {
			log.Info().
				Int(util.Quantity, simulation.EvenLen()).
				Str(util.Sigma, "$0.000").
				Str(util.Hyperlink, resultUrl(productID, "evn", port())).
				Msg(util.Sim + " ... " + fmt.Sprintf("%4s", util.NoTrend))
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

		n := simulation.TotalAfterFees() * size
		v := simulation.TotalEntries() * size

		log.Info().
			Str(util.Sigma, util.Usd(n)).
			Str(util.Quantity, util.Usd(v)).
			Str("%", util.Money(simulation.Net()*size)).
			Msg(util.Sim + util.Break + fmt.Sprintf("%4s", simulation.symbol()))

		log.Info().Msg(util.Sim + " ..")

		winners += simulation.WonLen()
		losers += simulation.LostLen()
		trading += simulation.TradingLen()
		won += simulation.TotalWonAfterFees()
		lost += simulation.TotalLostAfterFees()
		net += n
		volume += v
		even += simulation.EvenLen()
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
	log.Info().Str("          %", util.Money((net/volume)*100)).Msg(util.Sim + " ...")
	log.Info().Msg(util.Sim + " ..")
	log.Info().Msg(util.Sim + " .")

	return newSite(results)
}

func resultUrl(productID, dir string, port int) string {
	return fmt.Sprintf("http://localhost:%d/%s/%s.html", port, productID, dir)
}

func getRates(ses *config.Session, productID string) ([]cbp.Rate, error) {

	log.Debug().Msg("get rates for " + productID)

	pg := db.NewDB()

	var r cbp.Rate
	if err := pg.AutoMigrate(r); err != nil {
		return nil, err
	}

	pg.Where("product_id = ?", productID).
		Order("unix desc").
		First(&r)

	var start time.Time
	if r != (cbp.Rate{}) {
		log.Debug().Msg("found previous rate found for " + productID)
		start = r.Time()
	} else {
		log.Debug().Msg("no previous rate found for " + productID)
		start = *ses.Start()
	}

	to := start.Add(time.Hour * 4)
	for {

		rates, err := cbp.GetHistoricRates(productID, start, to)
		if err != nil {
			return nil, err
		}

		for _, r := range rates {
			rc := cbp.NewRate(productID, r)
			pg.Create(&rc)
		}

		if to.After(time.Now()) {
			break
		}

		start = to
		to = to.Add(time.Hour * 4)
		log.Debug().Int("... building simulation data", len(rates)).Send()
	}

	var savedRates []cbp.Rate
	pg.Where("product_id = ?", productID).
		Where("unix >= ?", ses.Start().UnixNano()).
		Order("unix asc").
		Find(&savedRates)

	log.Debug().Msgf("got [%d] rates for [%s]", len(savedRates), productID)

	return savedRates, nil
}
