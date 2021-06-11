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
	if s.WonLen() > 0 {
		return util.ThumbsUp
	} else if s.EvenLen() > 0 {
		return util.NoTrend
	} else if s.LostLen() > 0 {
		return util.ThumbsDn
	} else if s.TotalTradingAfterFees() > 0 {
		return util.UpTrend
	}
	return util.DnTrend
}

func (s *simulation) directory() string {
	if s.WonLen() > 0 {
		return "won"
	} else if s.EvenLen() > 0 {
		return "evn"
	} else if s.LostLen() > 0 {
		return "lst"
	} else {
		return "dnf"
	}
}

func newSimulation(session *config.Session, productID string, rates []cbp.Rate, simulation *simulation) {

	simulation.productID = productID
	msg := util.Sim + util.Break + util.GetCurrency(productID) + util.Break

	var then, that cbp.Rate
	for i, this := range rates {

		if !session.InPeriod(this.Time()) {
			continue
		}

		if session.GetPattern(productID).MatchesTweezerBottomPattern(then, that, this) {

			chart := newChart(session, rates[i-2:], productID)
			if chart == nil {
				continue
			}

			if chart.isWinner() {
				log.Info().Msg(msg + util.ThumbsUp)
				simulation.Won = append(simulation.Won, *chart)
			} else if chart.isLoser() {
				log.Info().Msg(msg + util.ThumbsDn)
				simulation.Lost = append(simulation.Lost, *chart)
			} else if chart.isTrading() {
				s := util.UpTrend
				if chart.result() < 0 {
					s = util.DnTrend
				}
				log.Info().Msg(msg + s)
				simulation.Trading = append(simulation.Trading, *chart)
			} else if chart.isEven() {
				log.Info().Msg(msg + util.NoTrend)
				simulation.Even = append(simulation.Even, *chart)
			}
		}
		then = that
		that = this
	}
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
