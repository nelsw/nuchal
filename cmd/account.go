package cmd

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"nuchal/pkg/cmd/account"
	"nuchal/pkg/util"
	"os"
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
		Use:     "account --force-holds --recurring",
		Example: example,
		Run: func(cmd *cobra.Command, args []string) {

			forceHolds := cmd.Flag("force-holds").Value.String() == "true"
			recurring := cmd.Flag("recurring").Value.String() == "true"

			run(forceHolds, recurring)

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {

				util.PrintCursor()

				command := scanner.Text()

				switch command {

				case "":
					fallthrough
				case "run":
					fmt.Printf("... running\n")
					run(forceHolds, false)

				case "exit":
					fmt.Printf("... exiting\n")
					os.Exit(0)

				default:
					fmt.Printf("... not familiar with the command [%s]\n", command)
					util.PrintCursor()
				}
			}
		}}

	c.Flags().Bool("force-holds", false, "If true, gain stops are placed to hold an entire balance.")
	c.Flags().Bool("recurring", false, "If true, audit will repeat every minute until the configured duration expires.")
	rootCmd.AddCommand(c)
}

func run(forceHolds, recurring bool) {
	if err := account.New(forceHolds, recurring); err != nil {
		log.Error().Err(err).Send()
	}
}
