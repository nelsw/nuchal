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
	"github.com/spf13/cobra"
)

var reportExample = `
	# Prints USD, Cryptocurrency, and total value of the configured Coinbase Pro account.
	# Also prints position and trading information, namely size, value, balance and holds.

	nuchal report
`

func init() {

	c := &cobra.Command{
		Use:     "report",
		Short:   "Provides a summary of your available currencies, balances, holds, and status of open trading positions",
		Example: reportExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := report.New(usd, size, gain, loss, delta); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(c)
}
