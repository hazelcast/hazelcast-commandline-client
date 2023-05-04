package viridian

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"math"
)

type CustomClassDeleteCmd struct{}

func (cmd CustomClassDeleteCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("delete-custom-class [file-name]")
	long := `Delete an existing custom class from the cluster.

Make sure you login before running this command.
`
	short := "Delete an existing custom class from the Viridian Cluster."
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, math.MaxInt)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")

	return nil
}

func (cmd CustomClassDeleteCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}

	cn := ec.Props().GetString("cluster.name")
	className := ec.Args()[0]

	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Deleting custom class")
		err := api.DeleteCustomClass(ctx, cn, className)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		ec.Logger().Error(err)
		return fmt.Errorf("error deleting custom class. Did you login?: %w", err)
	}
	stop()

	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("Custom class deleted successfully.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:delete-custom-class", &CustomClassDeleteCmd{}))
}
