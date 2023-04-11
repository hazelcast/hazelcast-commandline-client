//go:build base

package config

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/shell"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListCmd struct{}

func (cm ListCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	long := fmt.Sprintf(`Lists known configurations
	
A known configuration is a directory at %s that contains config.yaml.
Directory names which start with . or _ are ignored.
`, paths.Configs())
	short := "Lists known configurations"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm ListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	cd := paths.Configs()
	cs, err := config.FindAll(cd)
	if err != nil {
		ec.Logger().Warn("Cannot access configs directory at: %s: %s", cd, err.Error())
	}
	quite := ec.Props().GetBool(clc.PropertyQuiet) || shell.IsPipe()
	if len(cs) == 0 && !quite {
		I2(fmt.Fprintln(ec.Stderr(), "No configuration was found."))
		return nil
	}
	var rows []output.Row
	for _, c := range cs {
		rows = append(rows, output.Row{output.Column{
			Name:  "Config Name",
			Type:  serialization.TypeString,
			Value: c,
		}})
	}
	return ec.AddOutputRows(ctx, rows...)
}

func (ListCmd) Unwrappable() {}

func init() {
	Must(plug.Registry.RegisterCommand("config:list", &ListCmd{}))
}
