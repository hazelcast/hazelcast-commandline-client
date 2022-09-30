package versioncmd

import (
	"fmt"
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
			fmt.Printf("Hazelcast Command Line Client Version: %s\n", internal.ClientVersion)
			fmt.Printf("Latest Git Commit Hash: %s\n", internal.GitCommit)
			fmt.Printf("Hazelcast Go Client Version: %s\n", hazelcast.ClientVersion)
			fmt.Printf("Go Version: %s\n", runtime.Version())
		},
	}
	return &cmd
}
