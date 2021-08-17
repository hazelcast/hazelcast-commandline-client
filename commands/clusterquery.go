package commands

import (
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var clusterQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "retrieve information from the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		defer internal.HzDefer()
		fmt.Println(internal.ClusterConnect(cmd, "query", nil))
	},
}

func init() {
	clusterCmd.AddCommand(clusterQueryCmd)
}
