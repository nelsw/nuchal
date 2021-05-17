package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nchl/pkg/cmd/trade"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "trade",
		Short: "run trades",
		Long:  `run trades from a predefined strategy`,
		Run: func(cmd *cobra.Command, args []string) {

			if err := trade.New(); err != nil {
				log.Error().Err(err)
				panic(err)
			}

		}})
}
