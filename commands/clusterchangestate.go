package commands

import (
	"fmt"
	"log"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var (
	newState              string
	clusterChangeStateCmd = &cobra.Command{
		Use:   "change-state [--state new-state]",
		Short: "change state of the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			defer internal.ErrorRecover()
			config, err := internal.MakeConfig(cmd)
			if err != nil {
				log.Fatal(err)
			}
			result, err := internal.CallClusterOperation(config, "change-state", &newState)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(*result)
		},
	}
)

func init() {
	clusterCmd.PersistentFlags().StringVarP(&newState, "state", "s", "", "new state of the cluster")
	clusterCmd.MarkFlagRequired("state")
	clusterCmd.RegisterFlagCompletionFunc("state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"active", "no_migration", "frozen", "passive"}, cobra.ShellCompDirectiveDefault
	})
	clusterCmd.AddCommand(clusterChangeStateCmd)
}
