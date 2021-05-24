package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nuchal/pkg/cmd/report"
	"nuchal/pkg/util"
)

var reportExample = `
	# Print report report stats.
	nuchal report

	# Print report report stats for a user.
	nuchal report --user 'Carl Brutananadilewski' 

	# Print report report stats, every minute.
	nuchal report --recurring

	# Print report report stats, and place limit orders to hold the full balance.
	nuchal report --force-holds`

func init() {

	c := &cobra.Command{
		Use:     "report --force-holds --recurring --user",
		Example: reportExample,
		Run: func(cmd *cobra.Command, args []string) {

			forceHolds := cmd.Flag("force-holds").Value.String() == "true"
			recurring := cmd.Flag("recurring").Value.String() == "true"
			user := cmd.Flag("user").Value.String()

			if err := report.New(user, forceHolds, recurring); err != nil {
				log.Error().Err(err).Send()
			}
			util.LogBanner()
		},
	}

	c.Flags().String("user", util.GuestName,
		"Name of the user for simulating trades.")

	c.Flags().Bool("force-holds", false,
		"If true, gain stops are placed to hold an entire balance.")

	c.Flags().Bool("recurring", false,
		"If true, audit will repeat every minute until the configured duration expires.")

	rootCmd.AddCommand(c)
}
