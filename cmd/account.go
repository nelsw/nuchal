package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nuchal/pkg/cmd/account"
)

var (
	use     = "account --force-holds --recurring"
	example = `
	# Print account account stats.
	nuchal account

	# Print account account stats, every minute.
	nuchal account --recurring

	# Print account account stats, and place limit orders to hold the full balance.
	nuchal account --force-holds`
)

func init() {

	c := &cobra.Command{
		Use:     use,
		Example: example,
		Run:     run,
	}

	c.Flags().Bool("force-holds", false,
		"If true, gain stops are placed to hold an entire balance.")

	c.Flags().Bool("recurring", false,
		"If true, audit will repeat every minute until the configured duration expires.")

	rootCmd.AddCommand(c)
}

func run(cmd *cobra.Command, args []string) {

	forceHolds := cmd.Flag("force-holds").Value.String() == "true"
	recurring := cmd.Flag("recurring").Value.String() == "true"

	if err := account.New(forceHolds, recurring); err != nil {
		log.Error().Err(err).Send()
	}
}
