package mapcmd

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
)

func NewClear(config *hazelcast.Config) (*cobra.Command, error) {
	var mapName string
	cmd := &cobra.Command{
		Use:   "clear [--name mapname]",
		Short: "Clear entries of specified map",
		RunE: func(cmd *cobra.Command, args []string) error {
			// context timeout can be given according to bulk size of operation
			// we assume that current payload is same for all hazelcast operations
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			var err error
			defer cancel()
			m, err := getMap(ctx, config, mapName)
			if err != nil {
				return err
			}
			err = m.Clear(ctx)
			if err != nil {
				var handled bool
				handled, err = cloudcb(err, config)
				if handled {
					return err
				}
				return hzcerror.NewLoggableError(err, "Cannot clear map %s", mapName)
			}
			return nil
		},
	}
	if err := decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name"); err != nil {
		return nil, err
	}
	return cmd, nil
}
