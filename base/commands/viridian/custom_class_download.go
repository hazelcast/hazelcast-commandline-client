package viridian

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"strconv"
)

const outputPath = "output"

type CustomClassDownloadCmd struct{}

func (cmd CustomClassDownloadCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("download-custom-class [file-name]")
	long := `Download an existing custom class from the specified Viridian Cluster.

Make sure you login before running this command.
`
	short := "Download an existing custom class from the Viridian Cluster."
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(2, 2)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(outputPath, "o", "", false, "Download Path")

	return nil
}

func (cmd CustomClassDownloadCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}

	clusterName := ec.Args()[0]
	artifact := ec.Args()[1]
	artifactID, err := strconv.ParseInt(artifact, 10, 64)
	if err != nil {
		return err
	}
	path := ec.Props().GetString(outputPath)

	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Downloading custom class")
		err = api.DownloadCustomClass(ctx, sp, clusterName, artifactID, path)
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
