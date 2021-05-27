package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/trade"
	"github.com/spf13/cobra"
)

var tradeExample = `
	
	# Trade all products configured in pkg/config/patterns.json.
	# Trading creates active trading positions, AKA an available balance.
	# This will run until it uses all cash available.
	# This function is not perfect, not even graceful.

	nuchal trade

	# Hold the available balance for all products configured in pkg/config/patterns.json.	
	# Holding means to place a limit entry order on every active trading position.
	# This is best if you want to set stop gains and shut down.
	# This is the safest function.

	nuchal trade --hold

	# Sell the available balance for all products configured in pkg/config/patterns.json.
	# Selling means to place a limit loss order on on every active trading position that meet or exceeds the goal.
	# Selling will also attempt to break even after 45 minutes of trading.
	# This is an experimental function.

	nuchal trade --sell

	# Exit the available balance for all products configured in pkg/config/patterns.json.
	# Exiting means to place a market entry order on ever active trading position.
	# Exiting will effectively fire sale your open positions at the market price.
	# This is a safe function.
	
	nuchal trade --exit

`

func init() {

	c := &cobra.Command{
		Use:     "trade",
		Example: tradeExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := trade.New(); err != nil {
				panic(err)
			}
		}}

	rootCmd.AddCommand(c)
}
