package mapcmd

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	flagdecs "github.com/hazelcast/hazelcast-commandline-client/types/flagdecorators"
)

func NewRemove(config *hazelcast.Config) *cobra.Command {
	var all bool
	var (
		mapName,
		mapKey string
	)
	isValid := func() error {
		switch {
		case mapKey == "" && !all:
			return hzcerror.NewLoggableError(nil, "missing flags or arguments")
		case mapKey != "" && all:
			return hzcerror.NewLoggableError(nil, "given flags and arguments overlap with each other")
		}
		return nil
	}
	cloudcb := func(err error, config *hazelcast.Config) error {
		isCloudCluster := config.Cluster.Cloud.Enabled
		if networkErrMsg, handled := internal.TranslateNetworkError(err, isCloudCluster); handled {
			err = hzcerror.NewLoggableError(err, networkErrMsg)
		}
		return err
	}
	cmd := &cobra.Command{
		Use:   "remove --name mapname {--key keyname, --all}",
		Short: "Remove key(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if err = isValid(); err != nil {
				return err
			}
			// same context timeout for both single entry removal and map cleanup
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			if all {
				if err = m.Clear(ctx); err != nil {
					cmd.Printf("Cannot remove all entries %s\n", mapName)
					return cloudcb(err, config)
				}
				return nil
			}
			_, err = m.Remove(ctx, mapKey)
			if err != nil {
				cmd.Printf("Cannot remove key %s from the map %s\n", mapKey, mapName)
				return cloudcb(err, config)
			}
			return nil
		},
	}
	flagdecs.DecorateCommandWithAllFlag(cmd, &all, false, "refer all entries of the map")
	decorateCommandWithKeyFlagsNotRequired(cmd, &mapKey)
	decorateCommandWithMapNameFlags(cmd, &mapName)
	return cmd
}
