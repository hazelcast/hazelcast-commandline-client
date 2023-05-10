package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type CustomClassUploadCmd struct{}

func (cmd CustomClassUploadCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("upload-custom-class [cluster-ID/name] [path-to-class] [flags]")
	long := `Upload a new Custom Class to the specified Viridian Cluster.

Make sure you login before running this command.
`
	short := "Upload a Custom Class to the specified Viridian Cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(2, 2)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (cmd CustomClassUploadCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	cn := ec.Args()[0]
	filePath := ec.Args()[1]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Uploading custom class")
		err := api.UploadCustomClasses(ctx, sp, cn, filePath)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		ec.Logger().Error(err)
		return fmt.Errorf("error uploading custom classes. Did you login?: %w", err)
	}
	stop()
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("Custom class uploaded successfully.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:upload-custom-class", &CustomClassUploadCmd{}))
}
