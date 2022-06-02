package mapcmd

import (
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const MapUseExample = "map use m1    # sets the default map name to m1 unless set explicitly"

func NewUse() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `use [map-name | --reset]`,
		Short: "sets default map name",
		Example: MapUseExample + `
map get --key k1    # --name m1\" is inferred
map use --reset	    # resets the behaviour`,
		RunE: func(cmd *cobra.Command, args []string) error {
			persister := internal.PersistedNamesFromContext(cmd.Context())
			if cmd.Flags().Changed("reset") {
				delete(persister, "map")
				return nil
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			if len(args) > 1 {
				cmd.Println("Provide map name between \"\" quotes if it contains white space")
				return nil
			}
			persister["map"] = args[0]
			return nil
		},
	}
	_ = cmd.Flags().BoolP(MapResetFlag, "", false, "unset default name for map")
	return cmd
}
