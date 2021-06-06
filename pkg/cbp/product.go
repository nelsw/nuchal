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
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"strconv"
)

// Product is an aggregate of a Coinbase Product and the type of pattern to apply towards trading the product.
type Product struct {
	cb.Product
	Pattern
}

func (p *Product) Url() string {
	return fmt.Sprintf(`https://pro.coinbase.com/trade/%s`, p.ID)
}

func (p *Product) NewMarketBuyOrder() *cb.Order {
	o := new(cb.Order)
	o.ProductID = p.Id
	o.Side = "buy"
	o.Size = p.size()
	o.Type = "market"
	return o
}

func (p *Product) NewMarketSellOrder(size string) *cb.Order {
	o := new(cb.Order)
	o.ProductID = p.Id
	o.Side = "sell"
	o.Size = size
	o.Type = "market"
	return o
}

func (p *Product) NewLimitSellEntryOrderAtGoalPrice(trade *Trade) *cb.Order {
	return p.NewLimitSellEntryOrder(p.GoalPrice(trade.Price()), trade.Fill.Size)
}

func (p *Product) NewLimitSellEntryOrder(price float64, size string) *cb.Order {
	o := new(cb.Order)
	o.Price = p.price(price)
	o.ProductID = p.Id
	o.Side = "sell"
	o.Size = size
	o.Stop = "entry"
	o.StopPrice = p.price(price)
	o.Type = "limit"
	return o
}

func (p *Product) NewLimitLossOrder(price float64, size string) *cb.Order {
	o := new(cb.Order)
	o.Price = p.price(price)
	o.ProductID = p.Id
	o.Side = "sell"
	o.Size = size
	o.Stop = "loss"
	o.StopPrice = p.price(price)
	o.Type = "limit"
	return o
}

func (p *Product) size() string {
	if p.Size == 0.0 {
		return p.BaseMinSize
	}
	result := strconv.FormatFloat(p.Size, 'f', -1, 64)
	return result
}

func (p *Product) price(f float64) string {
	//str := strconv.FormatFloat(f, 'f', -1, 64)
	//chunks := strings.Split(str, ".")
	//places := chunks[1]
	//length := len(places) -1
	//number := strconv.Itoa(length)
	//result := fmt.Sprintf("%." +number+ "f", f)
	return fmt.Sprintf("%.3f", f)
}
