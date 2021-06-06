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
)

type simulation struct {

	// Product is an aggregate of the product to trade, and the pattern which used to trade.
	cbp.Product

	// Won are charts where we were profitable or broke even.
	Won []Chart

	// Lost are charts where we were not profitable.
	Lost []Chart

	// Trading are charts that never completed the simulation, they are still trading.
	Trading []Chart

	// Even are charts that broke even, not bad.
	Even []Chart
}

func (s *simulation) symbol() string {
	if s.Total() > 0 {
		return util.Won
	} else if s.Total() == 0 {
		return util.Even
	} else {
		return util.Lost
	}
}

func newSimulation(
	rates []cbp.Rate, product cbp.Product, maker, taker float64, period config.Period) *simulation {

	simulation := new(simulation)
	simulation.Product = product

	var then, that cbp.Rate
	for i, this := range rates {

		if !period.InPeriod(this.Time()) {
			continue
		}

		if product.MatchesTweezerBottomPattern(then, that, this) {

			chart := newChart(maker, taker, rates[i-2:], product)
			if chart.isWinner() {
				simulation.Won = append(simulation.Won, *chart)
			} else if chart.isLoser() {
				simulation.Lost = append(simulation.Lost, *chart)
			} else if chart.isTrading() {
				simulation.Trading = append(simulation.Trading, *chart)
			} else if chart.isEven() {
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

func (s *simulation) WonSum() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.result()
	}
	return sum
}

func (s *simulation) LostSum() float64 {
	sum := 0.0
	for _, l := range s.Lost {
		sum += l.result()
	}
	return sum
}

func (s *simulation) TradingSum() float64 {
	sum := 0.0
	for _, l := range s.Trading {
		sum += l.result()
	}
	return sum
}

func (s *simulation) Volume() float64 {
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
	return sum * s.Size
}

func (s *simulation) Total() float64 {
	return s.WonSum() + s.LostSum() + s.TradingSum()
}

func (s *simulation) Net() float64 {
	return s.Total() / s.Volume() * 100
}
