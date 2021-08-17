package commands

import (
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var (
	newState              string
	clusterChangeStateCmd = &cobra.Command{
		Use:   "changestate",
		Short: "change state of the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			defer internal.HzDefer()
			fmt.Println(internal.ClusterConnect(cmd, "changestate", &newState))
		},
	}
)

func init() {
	clusterCmd.PersistentFlags().StringVarP(&newState, "state", "s", "", "new state of the cluster")
	clusterCmd.MarkFlagRequired("state")
	clusterCmd.AddCommand(clusterChangeStateCmd)
}
