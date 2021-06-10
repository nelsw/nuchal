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
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
)

type simulation struct {

	// Won are charts where we were profitable or broke even.
	Won []Chart

	// Lost are charts where we were not profitable.
	Lost []Chart

	// Trading are charts that never completed the simulation, they are still trading.
	Trading []Chart

	// Even are charts that broke even, not bad.
	Even []Chart

	productID string
}

func (s *simulation) symbol() string {
	if s.TotalAfterFees() > 0 {
		return util.Won
	} else if s.TotalAfterFees() == 0 {
		return util.Even
	} else {
		return util.Lost
	}
}

func newSimulation(session *config.Session, productID string, rates []cbp.Rate) *simulation {

	simulation := new(simulation)
	simulation.productID = productID

	var then, that cbp.Rate
	for i, this := range rates {

		if !session.InPeriod(this.Time()) {
			continue
		}

		if session.GetPattern(productID).MatchesTweezerBottomPattern(then, that, this) {

			chart := newChart(session, rates[i-2:], productID)
			if chart.isWinner() {
				log.Info().Msg(util.Sim + util.Break + util.GetCurrency(productID) + util.Break + "winner")
				simulation.Won = append(simulation.Won, *chart)
			} else if chart.isLoser() {
				log.Info().Msg(util.Sim + util.Break + util.GetCurrency(productID) + util.Break + "loser")
				simulation.Lost = append(simulation.Lost, *chart)
			} else if chart.isTrading() {
				log.Info().Msg(util.Sim + util.Break + util.GetCurrency(productID) + util.Break + "trading")
				simulation.Trading = append(simulation.Trading, *chart)
			} else if chart.isEven() {
				log.Info().Msg(util.Sim + util.Break + util.GetCurrency(productID) + util.Break + "broke even")
				simulation.Even = append(simulation.Even, *chart)
			}
		}
		then = that
		that = this
	}
	return simulation
}

func (s *simulation) WonLen() int {
	return len(s.Won)
}

func (s *simulation) LostLen() int {
	return len(s.Lost)
}

func (s *simulation) TradingLen() int {
	return len(s.Trading)
}

func (s *simulation) EvenLen() int {
	return len(s.Even)
}

func (s *simulation) TotalWonAfterFees() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.result()
	}
	return sum
}

func (s *simulation) TotalLostAfterFees() float64 {
	sum := 0.0
	for _, l := range s.Lost {
		sum += l.result()
	}
	return sum
}

func (s *simulation) TotalTradingAfterFees() float64 {
	sum := 0.0
	for _, l := range s.Trading {
		sum += l.result()
	}
	return sum
}

func (s *simulation) TotalEntries() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.Entry
	}
	for _, l := range s.Lost {
		sum += l.Entry
	}
	for _, e := range s.Trading {
		sum += e.Entry
	}
	return sum
}

func (s *simulation) TotalAfterFees() float64 {
	return s.TotalWonAfterFees() + s.TotalLostAfterFees() + s.TotalTradingAfterFees()
}

func (s *simulation) Net() float64 {
	return s.TotalAfterFees() / s.TotalEntries() * 100
}
