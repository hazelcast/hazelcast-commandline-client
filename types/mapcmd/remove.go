package mapcmd

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
)

func NewRemove(config *hazelcast.Config) (*cobra.Command, error) {
	var (
		mapName,
		mapKey string
	)
	cmd := &cobra.Command{
		Use:   "remove --name mapname {--key keyname | --all}",
		Short: "Remove key(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// same context timeout for both single entry removal and map cleanup
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			var err error
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			_, err = m.Remove(ctx, mapKey)
			if err != nil {
				var handled bool
				handled, err = cloudcb(err, config)
				if handled {
					return err
				}
				return hzcerror.NewLoggableError(err, "Cannot remove given key from map %s", mapName)
			}
			return nil
		},
	}
	if err := decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name"); err != nil {
		return nil, err
	}
	if err := decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry"); err != nil {
		return nil, err
	}
	return cmd, nil
}
