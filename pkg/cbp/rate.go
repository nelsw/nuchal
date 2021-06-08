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
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"time"
)

type Rate struct {
	Unix      int64  `json:"unix" gorm:"primaryKey"`
	ProductId string `json:"product_id" gorm:"primaryKey"`
	cb.HistoricRate
}

func NewRate(productID string, historicRate cb.HistoricRate) *Rate {
	rate := new(Rate)
	rate.Unix = historicRate.Time.UnixNano()
	rate.ProductId = productID
	rate.HistoricRate = historicRate
	return rate
}

func (v *Rate) IsDown() bool {
	return v.Open > v.Close
}

func (v *Rate) IsUp() bool {
	return !v.IsDown()
}

func (v *Rate) IsInit() bool {
	return v != nil && v != (&Rate{})
}

func (v *Rate) Time() time.Time {
	return time.Unix(0, v.Unix)
}

func (v *Rate) Label() string {
	return time.Unix(0, v.Unix).Format(time.Kitchen)
}

func (v *Rate) Data() [4]float64 {
	return [4]float64{v.Open, v.Close, v.Low, v.High}
}
