//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
)

type CustomClassDeleteCmd struct{}

func (cmd CustomClassDeleteCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("delete-custom-class")
	long := `Deletes a custom class from the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Deletes a custom class from the given Viridian cluster."
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the delete operation")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	cc.AddStringArg(argArtifactID, argTitleArtifactID)
	return nil
}

func (cmd CustomClassDeleteCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo("Custom class will be deleted irreversibly, proceed?")
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	// inputs
	cluster := ec.GetStringArg(argClusterID)
	artifact := ec.GetStringArg(argArtifactID)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Deleting custom class")
		err = api.DeleteCustomClass(ctx, cluster, artifact)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.PrintlnUnnecessary("OK Custom class was deleted.")
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("viridian:delete-custom-class", &CustomClassDeleteCmd{}))
}
