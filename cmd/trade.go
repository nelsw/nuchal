package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/trade"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var tradeExample = `
	# Trade from a predefined product strategy.
	nuchal trade`

func init() {

	c := &cobra.Command{
		Use:     "trade",
		Example: tradeExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := trade.New(); err != nil {
				log.Error().Err(err)
				panic(err)
			}
		}}

	rootCmd.AddCommand(c)
}
