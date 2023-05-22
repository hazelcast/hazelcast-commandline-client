package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type CustomClassUploadCmd struct{}

func (cmd CustomClassUploadCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("upload-custom-class [cluster-name/cluster-ID] [file-name] [flags]")
	long := `Uploads a new Custom Class to the specified Viridian cluster.

Make sure you login before running this command.
`
	short := "Uploads a Custom Class to the specified Viridian cluster"
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
		err := api.UploadCustomClasses(ctx, sp.SetProgress, cn, filePath)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.PrintlnUnnecessary("Custom class was uploaded.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:upload-custom-class", &CustomClassUploadCmd{}))
}
