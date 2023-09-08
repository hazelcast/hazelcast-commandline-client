//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	argPath      = "path"
	argTitlePath = "path"
)

type CustomClassUploadCmd struct{}

func (cmd CustomClassUploadCmd) Unwrappable() {}

func (cmd CustomClassUploadCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("upload-custom-class")
	long := `Uploads a new Custom Class to the specified Viridian cluster.

Make sure you login before running this command.
`
	short := "Uploads a Custom Class to the specified Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	cc.AddStringArg(argPath, argTitlePath)
	return nil
}

func (cmd CustomClassUploadCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	cn := ec.GetStringArg(argClusterID)
	path := ec.GetStringArg(argPath)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Uploading custom class")
		err := api.UploadCustomClasses(ctx, sp.SetProgress, cn, path)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.PrintlnUnnecessary("OK Custom class was uploaded.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:upload-custom-class", &CustomClassUploadCmd{}))
}
