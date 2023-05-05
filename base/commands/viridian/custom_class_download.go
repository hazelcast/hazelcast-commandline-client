package viridian

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"math"
)

type CustomClassDownloadCmd struct{}

func (cmd CustomClassDownloadCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("download-custom-class [file-name]")
	long := `Download an existing custom class from the cluster.

Make sure you login before running this command.
`
	short := "Download an existing custom class from the Viridian Cluster."
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, math.MaxInt)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")

	return nil
}

func (cmd CustomClassDownloadCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}

	cn := ec.Props().GetString("cluster.name")
	cn = "f0wuy8wg"
	className := ec.Args()[0]

	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Downloading custom class")
		err = api.DownloadCustomClass(ctx, sp, cn, className)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		ec.Logger().Error(err)
		return fmt.Errorf("error downloading custom class. Did you login?: %w", err)
	}
	stop()

	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("Custom class downloaded successfully.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:download-custom-class", &CustomClassDownloadCmd{}))
}
