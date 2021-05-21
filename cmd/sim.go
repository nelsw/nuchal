package cmd

import (
	"github.com/spf13/cobra"
	"nuchal/pkg/cmd/sim"
)

func init() {

	rootCmd.AddCommand(&cobra.Command{
		Use:   "sim",
		Short: "strategy simulation",
		Long:  `run a simulation of a predefined strategy`,
		Run: func(cmd *cobra.Command, args []string) {
			sim.New()
		}})

}
