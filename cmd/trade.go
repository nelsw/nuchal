package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nuchal/pkg/cmd/trade"
	"strings"
)

var tradeExample = `
	# Trade from a predefined product strategy.
	nuchal trade

	# Trade from a predefined product strategy for only the provided user.
	nuchal trade --user 'Carl Brutanandilewski'

	# rade from a predefined product strategy for the provided cryptocurrencies.
	nuchal trade --coin 'ADA,MATIC,XTZ'`

func init() {

	c := &cobra.Command{
		Use:     "trade --user --coin",
		Example: tradeExample,
		Run: func(cmd *cobra.Command, args []string) {

			user := cmd.Flag("user").Value.String()
			coinCsv := cmd.Flag("coin").Value.String()
			coins := strings.Split(coinCsv, ",")

			if err := trade.New(user, coins); err != nil {
				log.Error().Err(err)
				panic(err)
			}

		}}

	rootCmd.AddCommand(c)

	c.Flags().String("user", "Carl Brutanandilewski",
		"Name of the user for simulating trades.")

	c.Flags().StringArray("coin", []string{"ADA", "MATIC", "XTZ"},
		"A csv list of cryptocurrencies to trade.")

	rootCmd.AddCommand(c)
}
