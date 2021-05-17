package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nchl/pkg/cmd/status"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "group",
		Short: "group status",
		Long:  `get group status`,
		Run: func(cmd *cobra.Command, args []string) {

			if err := status.New(); err != nil {
				log.Error().Err(err)
				panic(err)
			}

		}})
}
