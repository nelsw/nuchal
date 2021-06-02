package cmd

import (
	"github.com/nelsw/nuchal/pkg/cmd/sim"
	"github.com/spf13/cobra"
)

var simExample = `

	# Prints a simulation result report and serves a local 
	# website for graphs of said simulation results.
	nuchal sim

`

func init() {

	c := &cobra.Command{
		Use:     "sim",
		Example: simExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := sim.New(); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(c)

}
