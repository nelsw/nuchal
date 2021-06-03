package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/trade"
	"github.com/spf13/cobra"
)

var (
	tradeExample = `
	
	# Trade, that is buy and sell configured products.
	# Trading creates active trading positions, AKA an available balance.
	# This will run until it uses all cash available.
	# This function is not perfect, not even graceful.

	nuchal trade

	# Hold the available balance for all configured products.	
	# Holding means to place a limit entry order on every active trading position.
	# This is best if you want to set stop gains and shut down.
	# This is the safest function.

	nuchal trade --hold

	# Sell the available balance for all configured products.
	# Selling means to place a limit loss order on on every active trading position that meet or exceeds the goal.
	# Selling will also attempt to break even after 45 minutes of trading.
	# This is an experimental function.

	nuchal trade --sell

	# Exit the available balance for all configured products.
	# Exiting means to place a market entry order on ever active trading position.
	# Exiting will effectively fire sale your open positions at the market price.
	# This is a safe function.
	
	nuchal trade --exit

`
	hold,
	sell,
	exit bool
	tradeCmd = &cobra.Command{
		Use:     "trade",
		Example: tradeExample,
		Run: func(cmd *cobra.Command, args []string) {

			var err error

			if hold {
				err = trade.NewHolds()
			} else if sell {
				err = trade.NewSells()
			} else if exit {
				err = trade.NewExits()
			} else {
				err = trade.New()
			}

			if err != nil {
				panic(err)
			}
		}}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&hold, "hold", "d", false, "Hold trades")
	rootCmd.PersistentFlags().BoolVarP(&sell, "sell", "s", false, "Sell trades")
	rootCmd.PersistentFlags().BoolVarP(&exit, "exit", "e", false, "Exit trades")
	rootCmd.AddCommand(tradeCmd)
}
