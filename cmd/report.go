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

package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/report"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/spf13/cobra"
)

func init() {

	c := new(cobra.Command)
	c.Use = "report"
	c.Long = util.Banner
	c.Short = "Provides a summary of your available currencies, balances, holds, and status of open trading positions"
	c.Example = `
	# Prints USD, Cryptocurrency, and total value of the configured Coinbase Pro account.
	# Also prints position and trading information, namely size, value, balance and holds.
	nuchal report`
	c.Run = func(cmd *cobra.Command, args []string) {
		if session, err := config.NewSession(cfg, usd, size, gain, loss, delta); err != nil {
			panic(err)
		} else if err := report.New(session); err != nil {
			panic(err)
		}
	}

	rootCmd.AddCommand(c)
}
