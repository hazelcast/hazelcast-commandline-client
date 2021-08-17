package commands

import (
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var clusterShutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "shuts down the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		defer internal.HzDefer()
		fmt.Println(internal.ClusterConnect(cmd, "shutdown", nil))
	},
}

func init() {
	clusterCmd.AddCommand(clusterShutdownCmd)
}
