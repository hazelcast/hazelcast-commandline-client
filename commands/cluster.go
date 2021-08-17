package commands

import (
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "administrative cluster operations",
	Long: `Administrative cluster operations which controls a 
	Hazelcast Cloud cluster by manipulating its state and other features.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
