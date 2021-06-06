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

package cbp

import (
	"github.com/rs/zerolog/log"
	"math"
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {

	// Id is concatenation of two currencies. eg. BTC-USD
	Id string `yaml:"id"`

	// Gain is a percentage used to produce the goal sell price from the entry buy price.
	Gain float64 `yaml:"gain"`

	// Loss is a percentage used to derive a limit sell price from the entry buy price.
	Loss float64 `yaml:"loss"`

	// Size is the amount of the transaction, using the products native quote increment.
	Size float64 `yaml:"size"`

	// Delta is the size of an acceptable difference between tweezer bottom candlesticks.
	Delta float64 `yaml:"delta"`
}

func NewPattern(size, gain, loss, delta float64) *Pattern {
	pattern := new(Pattern)
	pattern.Size = size
	pattern.Gain = gain
	pattern.Loss = loss
	pattern.Delta = delta
	return pattern
}

func (p *Pattern) GoalPrice(price float64) float64 {
	return price + (price * p.Gain)
}

func (p *Pattern) LossPrice(price float64) float64 {
	return price - (price * p.Loss)
}

func (p *Pattern) MatchesTweezerBottomPattern(then, that, this Rate) bool {
	return isTweezerBottomTrend(then, that, this) && isTweezerBottomValue(that, this, p.Delta)
}

func isTweezerBottomValue(u, v Rate, d float64) bool {
	f := math.Abs(math.Min(u.Low, u.Close) - math.Min(v.Low, v.Open))
	b := f <= d
	if b {
		log.Info().Str("product", v.ProductId).Float64("tweezer", d-f)
	}
	return b
}

func isTweezerBottomTrend(t, u, v Rate) bool {
	return t.IsInit() && u.IsInit() && t.IsDown() && u.IsDown() && v.IsUp()
}
