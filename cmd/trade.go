package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/trade"
	"github.com/spf13/cobra"
)

var tradeExample = `
	
	# Trade all the products configured in pkg/config/patterns.json.
	nuchal trade

	# Trade all the products configured in pkg/config/patterns.json that have orphaned fills.
	# An orphan fill is a market entry order fill that does not have an associated limit entry/limit order fill.
	# Associations are made by a margin of 
	# 
	# 	 goal.price  + (goal.price *  user.MakerFee * goal.price)
	#  - entry.price + (entry.price * user.takerFee * entry.price)
	# -------------------------------------------------------------
	#  = 
	nuchal trade --orphans

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
