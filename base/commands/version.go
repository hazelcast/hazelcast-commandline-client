package commands

import (
	"runtime"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type VersionCommand struct {
}

func (c VersionCommand) Init(ctx plug.CommandContext) error {
	return nil
}

func (c VersionCommand) Exec(ctx plug.ExecContext) error {
	so := ctx.Stdout()
	printf(so, "Hazelcast Command Line Client Version : %s\n", internal.ClientVersion)
	printf(so, "Latest Git Commit Hash                : %s\n", internal.GitCommit)
	printf(so, "Hazelcast Go Client Version           : %s\n", hazelcast.ClientVersion)
	printf(so, "Go Version                            : %s\n", runtime.Version())
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("version", &VersionCommand{}))
}
