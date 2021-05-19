package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nuchal/pkg/cmd/account"
)

var (
	example = `
	# Print account account stats
	./nuchal account
	
	# Print account account stats, 
    # and place limit orders to 
    # hold the full balance.
	./nuchal account --force-holds`
)

func init() {

	c := &cobra.Command{
		Use:     "account --force-holds",
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {
			v := cmd.Flag("force-holds").Value
			if err := account.NewWithForceHolds(v.String() == "true"); err != nil {
				log.Error().Err(err)
				panic(err)
			}
		}}

	c.Flags().Bool("force-holds", false, "If true, gain stops are placed to hold an entire balance.")
	rootCmd.AddCommand(c)
}
