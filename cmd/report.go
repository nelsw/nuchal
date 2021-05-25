package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/report"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var reportExample = `
	# Print report report stats.
	nuchal report

	# Print report report stats, every minute.
	nuchal report --recurring

	# Print report report stats, and place limit orders to hold the full balance.
	nuchal report --force-holds`

func init() {

	c := &cobra.Command{
		Use:     "report --force-holds --recurring",
		Example: reportExample,
		Run: func(cmd *cobra.Command, args []string) {

			forceHolds := cmd.Flag("force-holds").Value.String() == "true"
			recurring := cmd.Flag("recurring").Value.String() == "true"

			if err := report.New(forceHolds, recurring); err != nil {
				log.Error().Err(err).Send()
			}
			util.LogBanner()
		},
	}

	c.Flags().Bool("force-holds", false,
		"If true, gain stops are placed to hold an entire balance.")

	c.Flags().Bool("recurring", false,
		"If true, audit will repeat every minute until the configured duration expires.")

	rootCmd.AddCommand(c)
}
