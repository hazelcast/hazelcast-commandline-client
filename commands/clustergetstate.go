package commands

import (
	"fmt"
	"log"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var clusterGetStateCmd = &cobra.Command{
	Use:   "get-state",
	Short: "get state of the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		defer internal.ErrorRecover()
		config, err := internal.MakeConfig(cmd)
		if err != nil {
			log.Fatal(err)
		}
		result, err := internal.CallClusterOperation(config, "get-state", nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(*result)
	},
}

func init() {
	clusterCmd.AddCommand(clusterGetStateCmd)
}
