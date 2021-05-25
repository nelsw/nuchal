package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/sim"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var simExample = `
	# Print simulation result report.
	nuchal sim

	# Print simulation result report to the browser.
	nuchal sim --serve`

func init() {

	c := &cobra.Command{
		Use:     "sim --serve",
		Example: simExample,
		Run: func(cmd *cobra.Command, args []string) {

			serve := cmd.Flag("serve").Value.String() == "true"

			if err := sim.New(serve); err != nil {
				log.Error().Err(err).Send()
			}
		},
	}

	c.Flags().Bool("serve", false,
		"If true, will serve html depicting simulation results.")

	rootCmd.AddCommand(c)

}
