package commands

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

var mapPutCmd = &cobra.Command{
	Use:   "put",
	Short: "Put to map",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := getMap(nil, mapName)
		if err != nil {
			return fmt.Errorf("error getting map %s: %w", mapName, err)
		}
		if mapKey == "" {
			return errors.New("map key is required")
		}
		if mapValue == "" {
			return errors.New("map value is required")
		}
		// TODO: process returned value which is SerializedData
		_, err = m.Put(mapKey, mapValue)
		if err != nil {
			return fmt.Errorf("error putting value for key %s from map %s: %w", mapKey, mapName, err)
		}
		/*
			if value != nil {
				fmt.Println(value)
			}
		*/
		return nil
	},
}

func init() {
	mapPutCmd.PersistentFlags().StringVar(&mapKey, "key", "", "key of the map")
	mapPutCmd.PersistentFlags().StringVar(&mapValue, "value", "", "value of the map")
}
