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
	"github.com/nelsw/nuchal/pkg/util"
	cb "github.com/preichenberger/go-coinbasepro/v2"
)

type Trade struct {
	cb.Fill
}

func NewTrade(cbFill cb.Fill) *Trade {
	trade := new(Trade)
	trade.Fill = cbFill
	return trade
}

func (t Trade) Price() float64 {
	return util.Float64(t.Fill.Price)
}

func (t Trade) Size() float64 {
	return util.Float64(t.Fill.Size)
}

func (t Trade) Total() float64 {
	return t.Price() * t.Size()
}
