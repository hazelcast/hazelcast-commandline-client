package versioncmd

import (
	"runtime"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "Version and build information",
		Long:  `Version and build information including the Go version, Hazelcast Go Client version and latest Git commit hash.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Hazelcast Command Line Client Version: %s\n", internal.Version)
			cmd.Printf("Latest Git Commit Hash: %s\n", internal.GitCommit)
			cmd.Printf("Hazelcast Go Client Version: %s\n", hazelcast.ClientVersion)
			cmd.Printf("Go Version: %s\n", runtime.Version())
		},
	}
	return &cmd
}
