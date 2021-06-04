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
	"github.com/spf13/cobra"
)

var simExample = `
	# Prints a simulation result report and serve a local website to host graphical report analysis.

	nuchal sim`

func init() {

	c := &cobra.Command{
		Use:     "sim",
		Short:   "Evaluates product & pattern configuration through a mock trading session and interactive chart results",
		Example: simExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := sim.New(usd, size, gain, loss, delta); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(c)

}
