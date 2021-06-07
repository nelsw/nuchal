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

package test

import "github.com/nelsw/nuchal/pkg/config"

const (
	size  = 1.0
	gain  = .0195
	loss  = .495
	delta = .001
	cfg   = "config.yml"
)

var (
	usd = []string{}
)

func Session() *config.Session {
	session, err := config.NewSession(cfg, usd, size, gain, loss, delta)
	if err != nil {
		panic(err)
	}
	return session
}
