package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	clustercmd "github.com/hazelcast/hazelcast-commandline-client/commands/cluster"
	mapcmd "github.com/hazelcast/hazelcast-commandline-client/commands/types/map"
	"github.com/hazelcast/hazelcast-commandline-client/config"
)

// NewRoot initializes root command for non-interactive mode
func NewRoot() (*cobra.Command, *config.PersistentFlags) {
	var flags config.PersistentFlags
	root := &cobra.Command{
		Use:   "hzc {cluster | help | map} [--address address | --cloud-token token | --cluster-name name | --config config]",
		Short: "Hazelcast command-line client",
		Long:  "Hazelcast command-line client connects your command-line to a Hazelcast cluster",
		Example: "`hzc map --name my-map put --key hello --value world` - put entry into map directly\n" +
			"`hzc help` - print help",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	// This is used to generate completion scripts
	if mode := os.Getenv("MODE"); strings.EqualFold(mode, "dev") {
		root.CompletionOptions.DisableDefaultCmd = false
	}
	assignPersistentFlags(root, &flags)
	root.AddCommand(clustercmd.New(), mapcmd.New())
	return root, &flags
}

// assignPersistentFlags assigns top level flags to command
func assignPersistentFlags(cmd *cobra.Command, flags *config.PersistentFlags) {
	cmd.PersistentFlags().StringVarP(&flags.CfgFile, "config", "c", config.DefaultConfigPath(), fmt.Sprintf("config file, only supports yaml for now"))
	cmd.PersistentFlags().StringVarP(&flags.Address, "address", "a", "", fmt.Sprintf("addresses of the instances in the cluster (default is %s).", config.DefaultClusterAddress))
	cmd.PersistentFlags().StringVarP(&flags.Cluster, "cluster-name", "", "", fmt.Sprintf("name of the cluster that contains the instances (default is %s).", config.DefaultClusterName))
	cmd.PersistentFlags().StringVar(&flags.Token, "cloud-token", "", "your Hazelcast Cloud token.")
}
