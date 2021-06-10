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

import "sort"

type cull struct {
	ids *[]string
}

func (c cull) UsdSelectionProductIDs() []string {
	return *c.ids
}

func (c cull) IDS() []string {
	return *c.ids
}

func NewCull(usd, pat, all []string) *cull {

	c := new(cull)

	var ids []string
	if len(usd) > 0 {
		for _, id := range usd {
			ids = append(ids, id)
		}
	} else if len(pat) > 0 {
		for _, id := range pat {
			ids = append(ids, id)
		}
	} else {
		for _, id := range all {
			ids = append(ids, id)
		}
	}

	sort.Strings(ids)

	c.ids = &ids

	return c
}
