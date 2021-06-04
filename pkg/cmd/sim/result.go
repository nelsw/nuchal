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
)

type Simulation struct {

	// Product is an aggregate of the product to trade, and the pattern which used to trade.
	cbp.Product

	// Won are charts where we were profitable or broke even.
	Won []Chart

	// Lost are charts where we were not profitable.
	Lost []Chart

	// Ether are charts that never completed the simulation, these are bad.
	Ether []Chart

	// Even are charts that broke even, not bad.
	Even []Chart
}

func NewSimulation(rates []cbp.Rate, product cbp.Product, maker, taker float64, period config.Period) *Simulation {

	simulation := new(Simulation)
	simulation.Product = product

	var then, that cbp.Rate
	for i, this := range rates {

		if !period.InRange(this.Time()) {
			continue
		}

		if product.MatchesTweezerBottomPattern(then, that, this) {

			chart := NewChart(maker, taker, rates[i-2:], product)
			if chart.IsWinner() {
				simulation.Won = append(simulation.Won, *chart)
			} else if chart.IsLoser() {
				simulation.Lost = append(simulation.Lost, *chart)
			} else if chart.IsEther() {
				simulation.Ether = append(simulation.Ether, *chart)
			} else if chart.IsEven() {
				simulation.Even = append(simulation.Even, *chart)
			}
		}
		then = that
		that = this
	}
	return simulation
}

func (s *Simulation) WonLen() int {
	return len(s.Won)
}

func (s *Simulation) LostLen() int {
	return len(s.Lost)
}

func (s *Simulation) EtherLen() int {
	return len(s.Ether)
}

func (s *Simulation) EvenLen() int {
	return len(s.Even)
}

func (s *Simulation) WonSum() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.Result()
	}
	return sum
}

func (s *Simulation) LostSum() float64 {
	sum := 0.0
	for _, l := range s.Lost {
		sum += l.Result()
	}
	return sum
}

func (s *Simulation) Volume() float64 {
	sum := 0.0
	for _, w := range s.Won {
		sum += w.EntryPlusFee()
	}
	for _, l := range s.Lost {
		sum += l.EntryPlusFee()
	}
	for _, e := range s.Ether {
		sum += e.EntryPlusFee()
	}
	return sum * s.Size
}

func (s *Simulation) Total() float64 {
	return s.WonSum() + s.LostSum()
}

func (s *Simulation) Net() float64 {
	return s.Total() / s.Volume() * 100
}
