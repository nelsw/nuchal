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

package config

import (
	"github.com/nelsw/nuchal/pkg/cbp"
	"gopkg.in/yaml.v2"
	"os"
	"sort"
)

type ParagonConfig struct {
	Patterns []cbp.Pattern `yaml:"patterns"`
}

type paragon struct {
	patterns                map[string]cbp.Pattern
	size, gain, loss, delta float64
}

func NewParagon(name string, size, gain, loss, delta float64) *paragon {

	var err error
	var f *os.File

	c := new(ParagonConfig)

	p := new(paragon)
	p.size = size
	p.gain = gain
	p.loss = loss
	p.delta = delta
	p.patterns = map[string]cbp.Pattern{}

	if f, err = os.Open(name); err == nil {
		if err = yaml.NewDecoder(f).Decode(c); err == nil && c.isValid() {
			for _, pattern := range c.Patterns {
				pattern.InitPattern(size, gain, loss, delta)
				p.patterns[pattern.ID] = pattern
			}
			return p
		}
	}

	return p
}

func (p *paragon) GetPattern(productID string) *cbp.Pattern {
	if pattern, ok := p.patterns[productID]; ok {
		return &pattern
	}
	return &cbp.Pattern{
		productID,
		p.gain,
		p.loss,
		p.size,
		p.delta,
	}
}

func (p *ParagonConfig) isValid() bool {
	if p == nil {
		return false
	}
	return true
}

func (p *paragon) patternIDs() *[]string {
	var productIDs []string
	for productID := range p.patterns {
		productIDs = append(productIDs, productID)
	}
	sort.Strings(productIDs)
	return &productIDs
}
