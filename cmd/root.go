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
	"github.com/mitchellh/go-homedir"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "nuchal",
		Short: "An application for evaluating and executing systematic cryptocurrency trades.",
		Long:  util.Banner,
	}

	// cfg is the where the configuration file is located.
	cfg string

	// usd represents the the USD Products to command
	usd []string

	// size, gain, loss, and delta are global product pattern properties.
	// size is a factor applied to the minimum trade size, defining the actual trade size for creating orders.
	// gain is a factor applied to the trade purchase price, defining the goal price for making a gain.
	// loss is a factor applied to the trade purchase price, defining the limit price for taking a loss.
	// delta is a factor applied to the product quote increment, defining the trend proximity for matching a pattern.
	size, gain, loss, delta float64
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(func() {
		if cfg != "" {
			viper.SetConfigFile(cfg) // Use config file from the flag.
		} else {
			home, err := homedir.Dir() // Find home directory.
			cobra.CheckErr(err)
			viper.AddConfigPath(home)  // Search for a config in the home directory
			viper.AddConfigPath(".")   // Search for a config in the current directory
			viper.SetConfigType("yml") // Search for configs that end in yml
		}
		viper.AutomaticEnv() // read in environment variables that match
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	})
	rootCmd.PersistentFlags().StringArrayVar(&usd, "usd", nil, "scope of USD Products to command")
	rootCmd.PersistentFlags().Float64VarP(&size, "size", "q", 1, "minimum trade size")
	rootCmd.PersistentFlags().Float64VarP(&gain, "gain", "g", .0195, "trade gain goal")
	rootCmd.PersistentFlags().Float64VarP(&loss, "loss", "l", .0495, "trade loss limit")
	rootCmd.PersistentFlags().Float64VarP(&delta, "delta", "d", .001, "pattern similarity")
}
