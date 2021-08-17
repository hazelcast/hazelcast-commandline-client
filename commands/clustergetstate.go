package commands

import (
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var clusterGetStateCmd = &cobra.Command{
	Use:   "getstate",
	Short: "get state of the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		defer internal.HzDefer()
		fmt.Println(internal.ClusterConnect(cmd, "getstate", nil))
	},
}

func init() {
	clusterCmd.AddCommand(clusterGetStateCmd)
}
