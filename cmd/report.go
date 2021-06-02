package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/report"
	"github.com/spf13/cobra"
)

var reportExample = `

	# Prints USD, Cryptocurrency, and total value of the configured Coinbase Pro account.
	# Also prints position information, namely available balance and holds.
	nuchal report

`

func init() {

	c := &cobra.Command{
		Use:     "report",
		Example: reportExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := report.New(); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(c)
}
