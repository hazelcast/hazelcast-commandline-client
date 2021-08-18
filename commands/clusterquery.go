package commands

import (
	"fmt"
	"log"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var clusterQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "retrieve information from the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		defer internal.ErrorRecover()
		config, err := internal.MakeConfig(cmd)
		if err != nil {
			log.Fatal(err)
		}
		result, err := internal.CallClusterOperation(config, "query", nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(*result)
	},
}

func init() {
	clusterCmd.AddCommand(clusterQueryCmd)
}
