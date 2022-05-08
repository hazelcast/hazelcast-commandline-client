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
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
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
	MapOutputFlag         = "output"
)

type MapRequestedOutput int

const (
	MapOutputEntries MapRequestedOutput = 1 << iota
	MapOutputKeys
	MapOutputValues
)

func (m MapRequestedOutput) String() string {
	switch m {
	case MapOutputEntries:
		return "entries"
	case MapOutputKeys:
		return "keys"
	case MapOutputValues:
		return "values"
	default:
		return ""
	}
}

func New(config *hazelcast.Config) *cobra.Command {
	// context timeout for each map operation can be configurable
	var cmd = &cobra.Command{
		Use:   "map {get | put | remove} --name mapname --key keyname [--value-type type | --value-file file | --value value]",
		Short: "Map operations",
	}
	cmd.AddCommand(
		NewGet(config),
		NewPut(config),
		NewRemove(config))
	return cmd
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

func decorateCommandWithMapNameFlags(cmd *cobra.Command, mapName *string) {
	cmd.Flags().StringVarP(mapName, MapNameFlag, MapNameFlagShort, "", "specify the map name")
	cmd.MarkFlagRequired("name")
}

func decorateCommandWithKeyFlags(cmd *cobra.Command, mapKey *string) {
	cmd.Flags().StringVarP(mapKey, MapKeyFlag, MapKeyFlagShort, "", "key of the map")
	cmd.MarkFlagRequired("key")
}

func decorateCommandWithValueFlags(cmd *cobra.Command, mapValue, mapValueFile, mapValueType *string) {
	flags := cmd.Flags()
	flags.StringVarP(mapValueFile, MapValueFileFlag, MapValueFileFlagShort, "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
	flags.StringVarP(mapValueType, MapValueTypeFlag, MapValueTypeFlagShort, "string", "type of the value, one of: string, json")
	cmd.RegisterFlagCompletionFunc(MapValueTypeFlag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "string"}, cobra.ShellCompDirectiveDefault
	})
}

func decorateCommandWithKeyFlagsNotRequired(cmd *cobra.Command, mapKey *string) {
	cmd.Flags().StringVarP(mapKey, MapKeyFlag, MapKeyFlagShort, "", "key of the map")
}

func decorateCommandWithMapKeySliceFlags(cmd *cobra.Command, mapKeys *[]string, required bool, usage string) {
	cmd.Flags().StringSliceVarP(mapKeys, MapKeyFlag, MapKeyFlagShort, []string{}, usage)
	if required {
		cmd.MarkFlagRequired("keySlice")
	}
}

func decorateCommandWithMapValueSliceFlags(cmd *cobra.Command, mapValues *[]string, required bool, usage string) {
	cmd.Flags().StringSliceVarP(mapValues, MapValueFlag, MapValueFlagShort, []string{}, usage)
	if required {
		cmd.MarkFlagRequired("valueSlice")
	}
}

func decorateCommandWithMapValueFileSliceFlags(cmd *cobra.Command, mapValueFiles *[]string, required bool, usage string) {
	cmd.Flags().StringSliceVarP(mapValueFiles, MapValueFileFlag, MapValueFileFlagShort, []string{}, usage)
	if required {
		cmd.MarkFlagRequired("valueFileSlice")
	}
}

func decorateCommandWithMapValueTypeSliceFlags(cmd *cobra.Command, mapValueTypes *[]string, required bool, usage string) {
	cmd.Flags().StringSliceVarP(mapValueTypes, MapValueTypeFlag, MapValueTypeFlagShort, []string{}, usage)
	cmd.RegisterFlagCompletionFunc(MapValueTypeFlag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "string"}, cobra.ShellCompDirectiveDefault
	})
	if required {
		cmd.MarkFlagRequired("valueTypeSlice")
	}
}

func decorateCommandWithMapOutputFlag(cmd *cobra.Command, output *string, required bool, usage string) {
	cmd.Flags().StringVar(output, MapOutputFlag, MapOutputEntries.String(), usage)
	if required {
		cmd.MarkFlagRequired(MapOutputFlag)
	}
}
