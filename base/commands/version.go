//go:build std || version

package commands

import (
	"context"
	"fmt"
	"runtime"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type VersionCommand struct{}

func (VersionCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("version")
	help := "Print the version"
	cc.SetCommandHelp(help, help)
	return nil
}

func (VersionCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	if ec.Props().GetBool(clc.PropertyVerbose) {
		return ec.AddOutputRows(ctx,
			makeRow("Hazelcast CLC", internal.Version),
			makeRow("Latest Git Commit Hash", internal.GitCommit),
			makeRow("Hazelcast Go Client", hazelcast.ClientVersion),
			makeRow("Go", fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)),
		)
	}
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Version",
			Type:  serialization.TypeString,
			Value: internal.Version,
		},
	})
}

func makeRow(key, value string) output.Row {
	return output.Row{
		output.Column{
			Name:  "Name",
			Type:  serialization.TypeString,
			Value: key,
		},
		output.Column{
			Name:  "Version",
			Type:  serialization.TypeString,
			Value: value,
		},
	}
}

func init() {
	check.Must(plug.Registry.RegisterCommand("version", &VersionCommand{}))
}
