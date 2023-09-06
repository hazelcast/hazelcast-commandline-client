//go:build std || version

package commands

import (
	"context"
	"fmt"
	"runtime"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type VersionCommand struct {
}

func (vc VersionCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("version")
	help := "Print the version"
	cc.SetCommandHelp(help, help)
	return nil
}

func (vc VersionCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	if ec.Props().GetBool(clc.PropertyVerbose) {
		return ec.AddOutputRows(ctx,
			vc.row("Hazelcast CLC", internal.Version),
			vc.row("Latest Git Commit Hash", internal.GitCommit),
			vc.row("Hazelcast Go Client", hazelcast.ClientVersion),
			vc.row("Go", fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)),
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

func (vc VersionCommand) row(key, value string) output.Row {
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

func (VersionCommand) Unwrappable() {}

func init() {
	Must(plug.Registry.RegisterCommand("version", &VersionCommand{}))
}
