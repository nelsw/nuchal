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
	"github.com/nelsw/nuchal/pkg/cmd/trade"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/spf13/cobra"
)

func init() {

	var drop, hold, sell, exit, eject bool

	c := new(cobra.Command)
	c.Use = "trade"
	c.Short = "Polls ticker data and executes buy & sell orders when conditions match product & pattern configuration."
	c.Long = util.Banner
	c.Example = `
  # Trade buys & sells products at prices or at times that meet or exceed pattern criteria, for a specified duration.
  nuchal trade

  # Hold creates a limit entry order at the goal price for every active trading position in your available balance.
  nuchal trade --hold

  # Sell all available positions (active trades) at prices or at times that meet or exceed pattern criteria.
  nuchal trade --sell

  # Sell all available positions (active trades) at the current market price. Will not sell holds.
  nuchal trade --exit

  # Drop will cancel every hold order, allowing the resulting products to be sold or converted.
  nuchal trade --drop

  # Sells everything at market price.
  nuchal trade --eject`

	c.Run = func(cmd *cobra.Command, args []string) {

		session, err := config.NewSession(cfg, usd, size, gain, loss, delta)
		if err != nil {
			panic(err)
		}

		if hold {
			err = trade.NewHolds(session)
		} else if sell {
			err = trade.NewSells(session)
		} else if exit {
			err = trade.NewExits(session)
		} else {
			err = trade.New(session)
		}

		if err != nil {
			panic(err)
		}
	}

	c.PersistentFlags().BoolVar(&hold, "hold", false, "Set a limit order for each trading position")
	c.PersistentFlags().BoolVar(&sell, "sell", false, "Close positions at the goal price or higher")
	c.PersistentFlags().BoolVar(&exit, "exit", false, "Liquidate all open positions at market price")
	c.PersistentFlags().BoolVar(&drop, "drop", false, "Cancel all hold orders to sell and convert")
	c.PersistentFlags().BoolVar(&eject, "eject", false, "Sells everything at market price")
	rootCmd.AddCommand(c)
}
