//go:build std || viridian

package viridian

import (
	"context"
	"errors"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type DownloadLogsCmd struct{}

func (cm DownloadLogsCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("download-logs")
	long := `Downloads the logs of the given Viridian cluster for the logged in API key.

Make sure you login before running this command.
`
	short := "Downloads the logs of the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagOutputDir, "o", "", false, "output directory for the log files, if not given current directory is used")
	cc.AddStringArg(argClusterID, "cluster name or ID", "specify the cluster name or ID")
	return nil
}

func (cm DownloadLogsCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.GetStringArg(argClusterID)
	outDir := ec.Props().GetString(flagOutputDir)
	// extract target info
	if err := validateOutputDir(outDir); err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Downloading cluster logs")
		err := api.DownloadClusterLogs(ctx, outDir, clusterNameOrID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:download-logs", &DownloadLogsCmd{}))
}

func validateOutputDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return nil
	}
	return errors.New("output-dir is not a directory")
}
