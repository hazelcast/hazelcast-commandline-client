package commands

import (
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
	help := "Print CLC version"
	cc.SetCommandHelp(help, help)
	return nil
}

func (vc VersionCommand) Exec(ec plug.ExecContext) error {
	if ec.Props().GetBool(clc.PropertyVerbose) {
		ec.AddOutputRows(
			vc.row("Hazelcast CLC", internal.Version),
			vc.row("Latest Git Commit Hash", internal.GitCommit),
			vc.row("Hazelcast Go Client", hazelcast.ClientVersion),
			vc.row("Go", fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)),
		)
		return nil
	}
	I2(fmt.Fprintln(ec.Stdout(), internal.Version))
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

func init() {
	Must(plug.Registry.RegisterCommand("version", &VersionCommand{}))
}
