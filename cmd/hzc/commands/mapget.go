package commands

import (
	"errors"
	"fmt"
	"github.com/alecthomas/chroma/quick"
	"github.com/hazelcast/hazelcast-go-client/v4/hazelcast"
	"github.com/hazelcast/hzc/cmd/hzc/commands/internal"
	"github.com/spf13/cobra"
	"os"
)

var mapGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get from map",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := getMap(internal.DefaultConfig(), mapName)
		if err != nil {
			return fmt.Errorf("error getting map %s: %w", mapName, err)
		}
		if mapKey == "" {
			return errors.New("map key is required")
		}
		value, err := m.Get(mapKey)
		if err != nil {
			return fmt.Errorf("error getting value for key %s from map %s: %w", mapKey, mapName, err)
		}
		if value != nil {
			switch v := value.(type) {
			case *hazelcast.JSONValue:
				if err := quick.Highlight(os.Stdout, v.ToString(), "json", "terminal", "tango"); err != nil {
					fmt.Println(v.ToString())
				}
			default:
				fmt.Println(value)
			}
		}
		return nil
	},
}

func init() {
	mapGetCmd.PersistentFlags().StringVar(&mapKey, "key", "", "key of the map")
}
