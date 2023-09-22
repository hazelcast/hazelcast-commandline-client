//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type CustomClassListCommand struct{}

func (CustomClassListCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list-custom-classes")
	long := `Lists all custom classes in the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Lists all custom classes in the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	return nil
}

func (CustomClassListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	cmd.IncrementMetric(ctx, ec, "total.viridian")
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	cn := ec.GetStringArg(argClusterID)
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	csi, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Retrieving custom classes")
		cs, err := api.ListCustomClasses(ctx, cn)
		if err != nil {
			return nil, err
		}
		return cs, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	cs := csi.([]viridian.CustomClass)
	if len(cs) == 0 {
		ec.PrintlnUnnecessary("OK There are no custom classes on this cluster.")
		return nil
	}
	rows := make([]output.Row, len(cs))
	for i, c := range cs {
		r := output.Row{
			output.Column{
				Name:  "ID",
				Type:  serialization.TypeInt64,
				Value: c.ID,
			},
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: c.Name,
			},
			output.Column{
				Name:  "Generated File Name",
				Type:  serialization.TypeString,
				Value: c.GeneratedFilename,
			},
			output.Column{
				Name:  "Status",
				Type:  serialization.TypeString,
				Value: c.Status,
			},
		}
		if verbose {
			r = append(r, output.Column{
				Name:  "Temporary Custom Classes ID",
				Type:  serialization.TypeString,
				Value: c.TemporaryCustomClassesId,
			})
		}
		rows[i] = r
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:list-custom-classes", &CustomClassListCommand{}))
}
