//go:build std || viridian

package viridian

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type DownloadLogsCommand struct{}

func (DownloadLogsCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("download-logs")
	long := `Downloads the logs of the given Viridian cluster for the logged in API key.

Make sure you login before running this command.
`
	short := "Downloads the logs of the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagOutputDir, "o", ".", false, "output directory for the log files; current directory is used by default")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	return nil
}

func (DownloadLogsCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.Metrics().Increment(metrics.NewSimpleKey(), "total.viridian."+cmd.RunningMode(ec))
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterNameOrID := ec.GetStringArg(argClusterID)
	outDir := ec.Props().GetString(flagOutputDir)
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return err
	}
	if err := validateOutputDir(outDir); err != nil {
		return err
	}
	st := stage.Stage[string]{
		ProgressMsg: "Downloading the cluster logs",
		SuccessMsg:  "Downloaded the cluster logs",
		FailureMsg:  "Failed downloading the cluster logs",
		Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
			return api.DownloadClusterLogs(ctx, outDir, clusterNameOrID)
		},
	}
	dir, err := stage.Execute(ctx, ec, "", stage.NewFixedProvider(st))
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Directory",
			Type:  serialization.TypeString,
			Value: dir,
		},
	})
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
	return fmt.Errorf("not a directory: %s", dir)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:download-logs", &DownloadLogsCommand{}))
}
