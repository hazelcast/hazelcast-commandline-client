/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package mapcmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

const (
	MapNameFlagShort      = "n"
	MapNameFlag           = "name"
	MapKeyFlagShort       = "k"
	MapKeyFlag            = "key"
	MapValueFlagShort     = "v"
	MapValueFlag          = "value"
	MapValueFileFlagShort = "f"
	MapValueFileFlag      = "value-file"
	MapValueTypeFlagShort = "t"
	MapValueTypeFlag      = "value-type"
)

func New(config *hazelcast.Config) (*cobra.Command, error) {
	// context timeout for each map operation can be configurable
	var cmd = &cobra.Command{
		Use:   "map {get | put | remove} --name mapname --key keyname [--value-type type | --value-file file | --value value]",
		Short: "Map operations",
	}
	if err := registerToMap(cmd, config); err != nil {
		return nil, err
	}
	return cmd, nil
}

func registerToMap(cmd *cobra.Command, config *hazelcast.Config) error {
	var err error
	// add clear command
	var clear *cobra.Command
	if clear, err = NewClear(config); err != nil {
		return hzcerror.NewLoggableError(err, "cannot add command clear, internal error")
	}
	cmd.AddCommand(clear)
	// add get command
	var get *cobra.Command
	if get, err = NewGet(config); err != nil {
		return hzcerror.NewLoggableError(err, "cannot add command get, internal error")
	}
	cmd.AddCommand(get)
	// add get-all command
	var getAll *cobra.Command
	if getAll, err = NewGetAll(config); err != nil {
		return hzcerror.NewLoggableError(err, "cannot add command get-all, internal error")
	}
	cmd.AddCommand(getAll)
	// add put command
	var put *cobra.Command
	if put, err = NewPut(config); err != nil {
		return hzcerror.NewLoggableError(err, "cannot add command put, internal error")
	}
	cmd.AddCommand(put)
	// add put-all command
	var putAll *cobra.Command
	if putAll, err = NewPutAll(config); err != nil {
		return hzcerror.NewLoggableError(err, "cannot add command put-all, internal error")
	}
	cmd.AddCommand(putAll)
	// add remove command
	var remove *cobra.Command
	if remove, err = NewRemove(config); err != nil {
		return hzcerror.NewLoggableError(err, "cannot add command remove, internal error")
	}
	cmd.AddCommand(remove)
	return nil
}

func normalizeMapValue(v, vFile, vType string) (interface{}, error) {
	var valueStr string
	var err error
	switch {
	case v != "" && vFile != "":
		return nil, hzcerror.NewLoggableError(nil, "Only one of --value and --value-file must be specified")
	case v != "":
		valueStr = v
	case vFile != "":
		if valueStr, err = loadValueFile(vFile); err != nil {
			err = hzcerror.NewLoggableError(err, "Cannot load the value file. Make sure file exists and process has correct access rights")
		}
	default:
		err = hzcerror.NewLoggableError(nil, "One of the value flags (--value or --value-file) must be set")
	}
	if err != nil {
		return nil, err
	}
	switch vType {
	case internal.TypeString:
		return valueStr, nil
	case internal.TypeJSON:
		return serialization.JSON(valueStr), nil
	}
	return nil, hzcerror.NewLoggableError(nil, "Provided value type parameter (%s) is not a known type. Provide either 'string' or 'json'", vType)
}

func loadValueFile(path string) (string, error) {
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

func cloudcb(err error, config *hazelcast.Config) (bool, error) {
	isCloudCluster := config.Cluster.Cloud.Enabled
	if networkErrMsg, handled := internal.TranslateNetworkError(err, isCloudCluster); handled {
		err = hzcerror.NewLoggableError(err, networkErrMsg)
		return true, err
	}
	return false, err
}

func withDashPrefix(flag string, short bool) string {
	if flag == "" {
		return ""
	}
	if short {
		return fmt.Sprintf("-%s", flag)
	}
	return fmt.Sprintf("--%s", flag)
}

func ObtainOrderingOfValueFlags(args []string) (vOrder []byte, tOrder []int) {
	if len(args) == 0 {
		return
	}
	for _, arg := range args {
		vShort := withDashPrefix(MapValueFlagShort, true)
		v := withDashPrefix(MapValueFlag, false)
		fShort := withDashPrefix(MapValueFileFlagShort, true)
		f := withDashPrefix(MapValueFileFlag, false)
		tShort := withDashPrefix(MapValueTypeFlagShort, true)
		t := withDashPrefix(MapValueTypeFlag, false)
		switch arg {
		case vShort, v:
			vOrder = append(vOrder, 's')
		case fShort, f:
			vOrder = append(vOrder, 'f')
		case tShort, t:
			if len(vOrder) == 0 {
				return
			}
			tOrder = append(tOrder, len(vOrder)-1)
		}
	}
	return
}

func getMap(ctx context.Context, clientConfig *hazelcast.Config, mapName string) (result *hazelcast.Map, err error) {
	hzcClient, err := internal.ConnectToCluster(ctx, clientConfig)
	if err != nil {
		return nil, hzcerror.NewLoggableError(err, "Cannot get initialize client")
	}
	if result, err = hzcClient.GetMap(ctx, mapName); err != nil {
		if msg, isHandled := internal.TranslateNetworkError(err, clientConfig.Cluster.Cloud.Enabled); isHandled {
			err = hzcerror.NewLoggableError(err, msg)
		}
		return nil, err
	}
	return
}

func decorateCommandWithValueFlags(cmd *cobra.Command, mapValue, mapValueFile, mapValueType *string) error {
	flags := cmd.Flags()
	flags.StringVarP(mapValue, MapValueFlag, MapValueFlagShort, "", "value of the map")
	flags.StringVarP(mapValueFile, MapValueFileFlag, MapValueFileFlagShort, "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
	flags.StringVarP(mapValueType, MapValueTypeFlag, MapValueTypeFlagShort, "string", "type of the value, one of: string, json")
	err := cmd.RegisterFlagCompletionFunc(MapValueTypeFlag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "string"}, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		return err
	}
	return nil
}

func decorateCommandWithMapNameFlags(cmd *cobra.Command, mapName *string, required bool, usage string) error {
	cmd.Flags().StringVarP(mapName, MapNameFlag, MapNameFlagShort, "", usage)
	if required {
		if err := cmd.MarkFlagRequired(MapNameFlag); err != nil {
			return err
		}
	}
	return nil
}

func decorateCommandWithMapKeyFlags(cmd *cobra.Command, mapKey *string, required bool, usage string) error {
	cmd.Flags().StringVarP(mapKey, MapKeyFlag, MapKeyFlagShort, "", usage)
	if required {
		if err := cmd.MarkFlagRequired(MapKeyFlag); err != nil {
			return err
		}
	}
	return nil
}

func decorateCommandWithMapKeyArrayFlags(cmd *cobra.Command, mapKeys *[]string, required bool, usage string) error {
	cmd.Flags().StringArrayVarP(mapKeys, MapKeyFlag, MapKeyFlagShort, []string{}, usage)
	if required {
		if err := cmd.MarkFlagRequired(fmt.Sprintf("%s-array", MapKeyFlag)); err != nil {
			return err
		}
	}
	return nil
}

func decorateCommandWithMapValueArrayFlags(cmd *cobra.Command, mapValues *[]string, required bool, usage string) error {
	cmd.Flags().StringArrayVarP(mapValues, MapValueFlag, MapValueFlagShort, []string{}, usage)
	if required {
		if err := cmd.MarkFlagRequired(fmt.Sprintf("%s-array", MapValueFlag)); err != nil {
			return err
		}
	}
	return nil
}

func decorateCommandWithMapValueFileArrayFlags(cmd *cobra.Command, mapValueFiles *[]string, required bool, usage string) error {
	cmd.Flags().StringArrayVarP(mapValueFiles, MapValueFileFlag, MapValueFileFlagShort, []string{}, usage)
	if required {
		if err := cmd.MarkFlagRequired(fmt.Sprintf("%s-array", MapValueFileFlag)); err != nil {
			return err
		}
	}
	return nil
}

func decorateCommandWithMapValueTypeArrayFlags(cmd *cobra.Command, mapValueTypes *[]string, required bool, usage string) error {
	var err error
	cmd.Flags().StringArrayVarP(mapValueTypes, MapValueTypeFlag, MapValueTypeFlagShort, []string{}, usage)
	err = cmd.RegisterFlagCompletionFunc(MapValueTypeFlag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "string"}, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		return err
	}
	if required {
		if err = cmd.MarkFlagRequired(fmt.Sprintf("%s-array", MapValueTypeFlag)); err != nil {
			return err
		}
	}
	return nil
}
