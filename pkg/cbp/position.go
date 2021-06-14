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
	"sort"
)

type Position struct {
	cb.Account
	cb.Ticker
	buys,
	sells []Trade
}

func (p *Position) IsHeld() bool {
	return p.Balance() == p.hold()
}

func (p Position) Balance() float64 {
	return util.Float64(p.Account.Balance)
}

func (p Position) Value() float64 {
	return p.Price() * p.Balance()
}

func (p Position) Price() float64 {
	return util.Float64(p.Ticker.Price)
}

func (p Position) hold() float64 {
	return util.Float64(p.Account.Hold)
}

func NewPosition(account cb.Account, ticker cb.Ticker, fills []cb.Fill) *Position {

	p := new(Position)
	p.Account = account
	p.Ticker = ticker

	for _, fill := range fills {
		if fill.Side == "buy" {
			p.buys = append(p.buys, *NewTrade(fill))
		} else if fill.Side == "sell" {
			p.sells = append(p.sells, *NewTrade(fill))
		}
	}

	return p
}

func (p *Position) GetActiveTrades() []Trade {

	if p.IsHeld() {
		return nil
	}

	buys := p.buys
	sort.SliceStable(buys, func(i, j int) bool {
		return buys[i].CreatedAt.Time().After(buys[j].CreatedAt.Time())
	})

	var trading []Trade
	hold := p.hold()
	for _, trade := range buys {
		if hold >= p.Balance() {
			break
		}
		trading = append(trading, trade)
		hold += trade.Size()
	}

	sort.SliceStable(trading, func(i, j int) bool {
		return trading[i].CreatedAt.Time().Before(trading[j].CreatedAt.Time())
	})

	return trading
}
