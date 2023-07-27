//go:build std || version

package commands

import (
	"context"
	"fmt"
	"runtime"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
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
	help := "Print CLC version"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("version [flags]")
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
	if ec.Props().GetString(clc.PropertyFormat) == base.PrinterDelimited {
		I2(fmt.Fprintln(ec.Stdout(), internal.Version))
	} else {
		return ec.AddOutputRows(ctx, vc.row("Hazelcast CLC", internal.Version))
	}
	ec.Logger().Debugf("version command ran OK")
	return nil
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
