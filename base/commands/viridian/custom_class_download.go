//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const flagOutputPath = "output-path"

type CustomClassDownloadCommand struct{}

func (CustomClassDownloadCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("download-custom-class")
	long := `Downloads a custom class from the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Downloads a custom class from the given Viridian cluster."
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagOutputPath, "o", "", false, "download path")
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming overwrite")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	cc.AddStringArg(argArtifactID, argTitleArtifactID)
	return nil
}

func (CustomClassDownloadCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.Metrics().Increment(metrics.NewSimpleKey(), "total.viridian."+cmd.RunningModeString(ec))
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	// inputs
	clusterName := ec.GetStringArg(argClusterID)
	artifact := ec.GetStringArg(argArtifactID)
	target := ec.Props().GetString(flagOutputPath)
	// extract target info
	t, err := viridian.CreateTargetInfo(target)
	if err != nil {
		return err
	}
	// if it is an existing file, it means we will overwrite it if user confirms
	if t.IsOverwrite() {
		autoYes := ec.Props().GetBool(clc.FlagAutoYes)
		if !autoYes {
			p := prompt.New(ec.Stdin(), ec.Stdout())
			yes, err := p.YesNo("Given output file exists and it will be overwritten, proceed?")
			if err != nil {
				ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
				return errors.ErrUserCancelled
			}
			if !yes {
				return errors.ErrUserCancelled
			}
		}
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Downloading custom class")
		err = api.DownloadCustomClass(ctx, sp.SetProgress, t, clusterName, artifact)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.PrintlnUnnecessary("OK Custom class was saved.\n")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name: "Path",
			Type: serialization.TypeString,
			// TODO: t.Path should not have / as the suffix
			Value: t.Path + t.FileName,
		},
	})
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:download-custom-class", &CustomClassDownloadCommand{}))
}
