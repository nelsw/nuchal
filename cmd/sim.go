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
	"github.com/nelsw/nuchal/pkg/cmd/sim"
	"github.com/nelsw/nuchal/pkg/config"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/spf13/cobra"
)

func init() {

	var winnersOnly, noLosers bool

	c := new(cobra.Command)
	c.Use = "sim"
	c.Short = "Evaluates product & pattern configuration through a mock trading session and interactive chart results."
	c.Long = util.Banner
	c.Example = `
	# Prints a simulation result report and serves a local website to host graphical report analysis.
	nuchal sim

	# Prints a simulation result report where the net gain for each product simulation was greater than zero.
	nuchal sim -t --no-losers

	# Prints a simulation result report where the net gain for each product simulation was greater than zero and also 
	# where the amount of positions trading are zero.	
	nuchal sim -w --winners-only`

	c.Run = func(cmd *cobra.Command, args []string) {

		session, err := config.NewSession(cfg, dur, usd, size, gain, loss, delta, debug)
		if err != nil {
			panic(err)
		}

		if err := sim.New(session, winnersOnly, noLosers); err != nil {
			panic(err)
		}
	}

	c.PersistentFlags().BoolVarP(&winnersOnly, "winners-only", "w", false, "")
	c.PersistentFlags().BoolVarP(&noLosers, "no-losers", "t", false, "")
	rootCmd.AddCommand(c)
}
