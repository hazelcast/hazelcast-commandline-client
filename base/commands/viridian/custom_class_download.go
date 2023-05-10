package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const flagOutputPath = "output-path"

type CustomClassDownloadCmd struct{}

func (cmd CustomClassDownloadCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("download-custom-class [file-name/artifact-id] [flags]")
	long := `Download a custom class from the given Viridian Cluster.

Make sure you login before running this command.
`
	short := "Download a custom class from the given Viridian Cluster."
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(2, 2)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagOutputPath, "o", "", false, "download path")
	return nil
}

func (cmd CustomClassDownloadCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	// inputs
	clusterName := ec.Args()[0]
	artifact := ec.Args()[1]
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
			yes, err := p.YesNo("Such a file exists and it will be overwritten, proceed?")
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
		err = api.DownloadCustomClass(ctx, sp, t, clusterName, artifact)
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
