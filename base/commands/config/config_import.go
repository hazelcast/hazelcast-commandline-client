//go:build std || config

package config

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
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
	stages := config.MakeImportStages(ec, target)
	path, err := stage.Execute(ctx, ec, src, stage.NewFixedProvider(stages...))
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Configuration Path",
			Type:  serialization.TypeString,
			Value: path,
		},
	})
}

func init() {
	Must(plug.Registry.RegisterCommand("config:import", &ImportCmd{}))
}
