package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/sim"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var simExample = `
	# Print simulation result report.
	nuchal sim

	# Print simulation result report to the console, with user maker/taker fees.
	nuchal sim --user 'Carl Brutananadilewski'

	# Print simulation result report to the browser.
	nuchal sim --serve

	# Print simulation result report to console, with provided cryptocurrency symbols.
	nuchal sim --coin 'ADA,MATIC,XTZ'`

func init() {

	c := &cobra.Command{
		Use:     "sim --user --coin",
		Example: simExample,
		Run: func(cmd *cobra.Command, args []string) {

			user := cmd.Flag("user").Value.String()
			serve := cmd.Flag("serve").Value.String() == "true"

			if err := sim.New(user, serve); err != nil {
				log.Error().Err(err).Send()
			}
		},
	}

	c.Flags().String("user", util.GuestName,
		"Name of the user for simulating trades.")

	c.Flags().Bool("serve", false,
		"If true, will serve html depicting simulation results.")

	rootCmd.AddCommand(c)

}
