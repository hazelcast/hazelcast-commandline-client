package commands

import (
	"errors"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client/v4/hazelcast"
	"github.com/hazelcast/hzc/cmd/hzc/commands/internal"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var mapPutCmd = &cobra.Command{
	Use:   "put",
	Short: "Put to map",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var normalizedValue interface{}
		m, err := getMap(internal.DefaultConfig(), mapName)
		if err != nil {
			return fmt.Errorf("error getting map %s: %w", mapName, err)
		}
		if mapKey == "" {
			return internal.ErrMapKeyMissing
		}
		if normalizedValue, err = normalizeMapValue(); err != nil {
			return err
		}
		// TODO: process returned value which is SerializedData
		_, err = m.Put(mapKey, normalizedValue)
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

func normalizeMapValue() (interface{}, error) {
	// TODO: move flag related code out
	// --value and --value-file arguments are mutually exclusive
	var valueStr string
	var err error
	if mapValue != "" && mapValueFile != "" {
		return nil, internal.ErrMapValueAndFileMutuallyExclusive
	} else if mapValue != "" {
		valueStr = mapValue
	} else if mapValueFile != "" {
		if valueStr, err = loadValueFIle(mapValueFile); err != nil {
			return nil, fmt.Errorf("error loading value: %w", err)
		}
	} else {
		return nil, internal.ErrMapValueMissing
	}
	switch mapValueType {
	case internal.TypeString:
		return valueStr, nil
	case internal.TypeJSON:
		return hazelcast.CreateJSONValueFromString(valueStr), nil
	}
	return nil, fmt.Errorf("%s is not a known value type", mapValueType)
}

func loadValueFIle(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}
	if path == "-" {
		if value, err := ioutil.ReadAll(os.Stdin); err != nil {
			return "", err
		} else {
			return string(value), nil
		}
	}
	if value, err := ioutil.ReadFile(path); err != nil {
		return "", err
	} else {
		return string(value), nil
	}
}

func init() {
	decorateCommandWithKeyFlags(mapPutCmd)
	decorateCommandWithValueFlags(mapPutCmd)
}
