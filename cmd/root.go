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
		Short: "Evaluates & executes high frequency cryptocurrency trades from configurable trend alignment patterns.",
		Long:  util.Banner,
	}

	// debug is a flag to turn on debug logging
	debug bool

	// cfg is the where the configuration file is located.
	cfg string

	// dur is parsed by time.Duration to determine command or command data time frame
	dur string

	// usd represents the the USD Products to command
	usd []string

	// size, gain, loss, and delta are global product pattern properties.
	// size is a factor applied to the minimum trade size, defining the actual trade size for creating orders.
	// gain is a factor applied to the trade purchase price, defining the goal price for making a gain.
	// loss is a factor applied to the trade purchase price, defining the limit price for taking a loss.
	// delta is a value that defines the trend proximity for matching a pattern.
	size, gain, loss, delta float64
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(func() {

		viper.SetConfigFile(cfg)
		if err := viper.ReadInConfig(); err == nil {
			return
		}

		home, err := homedir.Dir() // Find home directory.
		cobra.CheckErr(err)

		cfg = home + "/nuchal.yml"
		viper.SetConfigFile(cfg)
		if err := viper.ReadInConfig(); err == nil {
			return
		}

		cfg = home + "/Desktop/nuchal.yml"
		viper.SetConfigFile(cfg)
		_ = viper.ReadInConfig()
	})
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "x", false, "debug mode")
	rootCmd.PersistentFlags().StringVarP(&dur, "duration", "p", "", "period duration")
	rootCmd.PersistentFlags().StringVarP(&cfg, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringSliceVar(&usd, "usd", nil, "scope of USD Products to command")
	rootCmd.PersistentFlags().Float64VarP(&size, "size", "q", 1, "minimum trade size")
	rootCmd.PersistentFlags().Float64VarP(&gain, "gain", "g", .0195, "trade gain goal")
	rootCmd.PersistentFlags().Float64VarP(&loss, "loss", "l", .195, "trade loss limit")
	rootCmd.PersistentFlags().Float64VarP(&delta, "delta", "d", .001, "pattern similarity")
}
