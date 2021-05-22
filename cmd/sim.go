package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nuchal/pkg/cmd/sim"
)

func init() {

	c := &cobra.Command{
		Use: "sim --user --coin",
		Example: `
	# Print simulation result report.
	nuchal sim

	# Print simulation result report, with user maker/taker fees.
	nuchal sim --user 'Carl Brutanandilewski'

	# Print simulation result report, with provided cryptocurrency symbols.
	nuchal sim --coin 'ADA,MATIC,XTZ'`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := sim.New(); err != nil {
				log.Error().Err(err).Send()
			}
		},
	}

	rootCmd.AddCommand(c)

}
