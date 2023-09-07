//go:build std || config

package config

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ImportCmd struct{}

func (cm ImportCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("import")
	short := "Imports configuration from an arbitrary source"
	long := `Imports configuration from an arbitrary source
	
Currently importing Viridian connection configuration is supported only.
	
1. On Viridian console, visit:
	
	Dashboard -> Connect Client -> CLI

2. Copy the URL in box 2 and pass it as the second parameter.
   Make sure the text is quoted before running:
	
	clc config import my-config "https://api.viridian.hazelcast.com/client_samples/download/..."
	
`
	cc.SetCommandHelp(long, short)
	cc.AddStringArg(argConfigName, argTitleConfigName)
	cc.AddStringArg(argSource, argTitleSource)
	return nil
}

func (cm ImportCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	target := ec.GetStringArg(argConfigName)
	src := ec.GetStringArg(argSource)
	path, err := config.ImportSource(ctx, ec, target, src)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("OK Created the configuration at: %s", path)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("config:import", &ImportCmd{}))
}
