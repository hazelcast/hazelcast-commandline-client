//go:build std || config

package config

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ImportCmd struct{}

func (cm ImportCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("import [configuration-name] [source] [flags]")
	short := "Imports configuration from an arbitrary source"
	long := `Imports configuration from an arbitrary source
	
Currently importing only Viridian connection configuration is supported.
	
1. On Viridian console, visit:
	
	Dashboard -> Connect Client -> Quick connection guide -> Python

2. Copy the text in box 1 and pass it as the second parameter.
   Make sure the text is quoted before running:
	
	clc config import my-config "curl https://api.viridian.hazelcast.com ... default.zip"
	
Alternatively, you can use an already downloaded Python client sample:
	
	clc config import my-config /home/me/Downloads/hazelcast-cloud-python-sample....zip
	
`
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(2, 2)
	return nil
}

func (cm ImportCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	target := ec.Args()[0]
	src := ec.Args()[1]
	path, err := config.ImportSource(ctx, ec, target, src)
	if err != nil {
		return err
	}
	if ec.Interactive() || ec.Props().GetBool(clc.PropertyVerbose) {
		I2(fmt.Fprintf(ec.Stdout(), "Created configuration at: %s\n", path))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("config:import", &ImportCmd{}))
}
