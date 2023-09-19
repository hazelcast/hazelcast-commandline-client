//go:build std || config

package config

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListCommand struct{}

func (ListCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	long := fmt.Sprintf(`Lists known configurations
	
A known configuration is a directory at %s that contains config.yaml.
Directory names which start with . or _ are ignored.
`, paths.Configs())
	short := "Lists known configurations"
	cc.SetCommandHelp(long, short)
	return nil
}

func (ListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	rows, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) ([]output.Row, error) {
		sp.SetText("Finding configurations")
		cd := paths.Configs()
		cs, err := config.FindAll(cd)
		if err != nil {
			ec.Logger().Warn("Cannot access configuration directory at: %s: %s", cd, err.Error())
		}
		var rows []output.Row
		for _, c := range cs {
			rows = append(rows, output.Row{output.Column{
				Name:  "Configuration Name",
				Type:  serialization.TypeString,
				Value: c,
			}})
		}
		return rows, nil
	})
	if err != nil {
		return err
	}
	stop()
	if len(rows) == 0 {
		ec.PrintlnUnnecessary("OK No configurations found.")
		return nil
	}
	msg := fmt.Sprintf("OK Found %d configurations.", len(rows))
	defer ec.PrintlnUnnecessary(msg)
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("config:list", &ListCommand{}))
}
