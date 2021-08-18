package commands

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/chroma/quick"
	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"
)

var mapGetCmd = &cobra.Command{
	Use:   "get [--name mapname | --key keyname]",
	Short: "Get from map",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.TODO()
		config, err := internal.MakeConfig(cmd)
		if err != nil {
			return err
		}
		m, err := getMap(config, mapName)
		if err != nil {
			return fmt.Errorf("error getting map %s: %w", mapName, err)
		}
		if mapKey == "" {
			return errors.New("map key is required")
		}
		value, err := m.Get(ctx, mapKey)
		if err != nil {
			return fmt.Errorf("error getting value for key %s from map %s: %w", mapKey, mapName, err)
		}
		if value != nil {
			switch v := value.(type) {
			case serialization.JSON:
				if err := quick.Highlight(os.Stdout, v.String(),
					"json", "terminal", "tango"); err != nil {
					fmt.Println(v.String())
				}
			default:
				fmt.Println(value)
			}
		}
		return nil
	},
}

func init() {
	mapGetCmd.PersistentFlags().StringVarP(&mapKey, "key", "k", "", "key of the map")
	mapGetCmd.MarkFlagRequired("key")
	mapCmd.AddCommand(mapGetCmd)
}
