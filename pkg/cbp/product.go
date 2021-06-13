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
	"gorm.io/gorm"
)

type Posture interface {
	ID() string
}

type Product struct {
	gorm.Model
	cb.Product
	Pattern
}

func (p *Product) ID() string {
	return p.Product.BaseCurrency + "-" + p.QuoteCurrency
}

func NewProduct(product cb.Product) Product {
	p := new(Product)
	p.Product = product
	return *p
}
