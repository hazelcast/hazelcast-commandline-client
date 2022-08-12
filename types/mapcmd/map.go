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
	"strings"
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
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
	MapKeyTypeFlag        = "key-type"
	MapResetFlag          = "reset"
)

func New(config *hazelcast.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "map {get | put | clear | put-all | get-all | remove} --name mapname --key keyname [--value-type type | --value-file file | --value value]",
		Short:   "Map operations",
		Example: fmt.Sprintf("%s\n%s\n%s", MapPutExample, MapGetExample, MapUseExample),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// All the following lines are to set map name if it is set by "use" command.
			// If the map name is given explicitly, do not set the one given with "use" command.
			// Missing flag errors are not handled here.
			// They are expected to be handled by the actual command.
			persister := internal.PersistedNamesFromContext(cmd.Context())
			val, isSet := persister["map"]
			if !isSet {
				return nil
			}
			nameFlag := cmd.Flag("name")
			if nameFlag == nil {
				// flag is absent
				return nil
			}
			if nameFlag.Changed {
				// flag value is set explicitly
				return nil
			}
			if err := cmd.Flags().Set("name", val); err != nil {
				return hzcerrors.NewLoggableError(err, "Default name for map cannot be set")
			}
			return nil
		},
	}
	cmd.AddCommand(
		NewPut(config),
		NewPutAll(config),
		NewGet(config),
		NewGetAll(config),
		NewRemove(config),
		NewClear(config),
		NewUse())
	return cmd
}

func isNegativeSecond(d time.Duration) error {
	if d.Seconds() <= 0 {
		return errors.New(fmt.Sprintf("duration %s must be positive", d))
	}
	return nil
}

func validateTTL(d time.Duration) error {
	if err := isNegativeSecond(d); err != nil {
		return err
	}
	// server side time resolution is one second
	fmt.Println(d.Seconds())
	if d.Seconds() < 1.0 {
		return errors.New("ttl duration cannot be less than a second")
	}
	return nil
}

func printValueBasedOnType(cmd *cobra.Command, value interface{}) {
	var err error
	switch v := value.(type) {
	case serialization.JSON:
		if err = quick.Highlight(cmd.OutOrStdout(), fmt.Sprintln(v),
			"json", "terminal", "tango"); err != nil {
			cmd.Println(v.String())
		}
	default:
		if v == nil {
			cmd.Println("There is no value corresponding to the provided key")
			break
		}
		cmd.Println(v)
	}
}

func normalizeMapValue(v, vFile, vType string) (interface{}, error) {
	var valueStr string
	var err error
	switch {
	case v != "" && vFile != "":
		return nil, hzcerrors.NewLoggableError(nil, "Only one of --value and --value-file must be specified")
	case v != "":
		valueStr = v
	case vFile != "":
		if valueStr, err = loadValueFile(vFile); err != nil {
			err = hzcerrors.NewLoggableError(err, "Cannot load the value file. Make sure file exists and process has correct access rights")
		}
	default:
		err = hzcerrors.NewLoggableError(nil, "One of the value flags (--value or --value-file) must be set")
	}
	if err != nil {
		return nil, err
	}
	mapValue, err := internal.ConvertString(valueStr, vType)
	if err != nil {
		err = hzcerrors.NewLoggableError(err, "Conversion error on value %s to value-type %s, %s", valueStr, vType, err)
	}
	return mapValue, err
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

func isCloudIssue(err error, config *hazelcast.Config) (bool, error) {
	isCloudCluster := config.Cluster.Cloud.Enabled
	if networkErrMsg, handled := hzcerrors.TranslateNetworkError(err, isCloudCluster); handled {
		err = hzcerrors.NewLoggableError(err, networkErrMsg)
		return true, err
	}
	return false, err
}

func withShortFlag(flag string) string {
	if flag == "" {
		panic("flag cannot be empty string")
	}
	return fmt.Sprintf("-%s", flag)
}

func withLongFlag(flag string) string {
	if flag == "" {
		panic("flag cannot be empty string")
	}
	return fmt.Sprintf("--%s", flag)
}

func ObtainOrderingOfValueFlags(args []string) (vOrder []byte) {
	if len(args) == 0 {
		return
	}
	for _, arg := range args {
		vShort := withShortFlag(MapValueFlagShort)
		v := withLongFlag(MapValueFlag)
		fShort := withShortFlag(MapValueFileFlagShort)
		f := withLongFlag(MapValueFileFlag)
		switch arg {
		case vShort, v:
			vOrder = append(vOrder, 's')
		case fShort, f:
			vOrder = append(vOrder, 'f')
		}
	}
	return
}

func getMap(ctx context.Context, clientConfig *hazelcast.Config, mapName string) (result *hazelcast.Map, err error) {
	hzcClient, err := internal.ConnectToCluster(ctx, clientConfig)
	if err != nil {
		return nil, hzcerrors.NewLoggableError(err, "Cannot get initialize client")
	}
	if result, err = hzcClient.GetMap(ctx, mapName); err != nil {
		if msg, isHandled := hzcerrors.TranslateNetworkError(err, clientConfig.Cluster.Cloud.Enabled); isHandled {
			err = hzcerrors.NewLoggableError(err, msg)
		}
		return nil, err
	}
	return
}

func decorateCommandWithValueFlags(cmd *cobra.Command, mapValue, mapValueFile *string) {
	flags := cmd.Flags()
	flags.StringVarP(mapValue, MapValueFlag, MapValueFlagShort, "", "value of the map")
	flags.StringVarP(mapValueFile, MapValueFileFlag, MapValueFileFlagShort, "", `path to the file that contains the value. Use "-" (dash) to read from stdin`)
}

func decorateCommandWithMapNameFlags(cmd *cobra.Command, mapName *string, required bool, usage string) {
	cmd.Flags().StringVarP(mapName, MapNameFlag, MapNameFlagShort, "", usage)
	if required {
		if err := cmd.MarkFlagRequired(MapNameFlag); err != nil {
			panic(err)
		}
	}
}

func decorateCommandWithMapKeyTypeFlags(cmd *cobra.Command, mapKeyType *string, required bool) {
	help := fmt.Sprintf("key type, one of: %s (default: string)", strings.Join(internal.SupportedTypeNames, ","))
	cmd.Flags().StringVar(mapKeyType, MapKeyTypeFlag, "", help)
	if required {
		if err := cmd.MarkFlagRequired(MapKeyTypeFlag); err != nil {
			panic(err)
		}
	}
	err := cmd.RegisterFlagCompletionFunc(MapKeyTypeFlag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return internal.SupportedTypeNames, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		panic(err)
	}
}

func decorateCommandWithMapValueTypeFlags(cmd *cobra.Command, mapValueType *string, required bool) {
	help := fmt.Sprintf("value type, one of: %s (default: string)", strings.Join(internal.SupportedTypeNames, ","))
	cmd.Flags().StringVarP(mapValueType, MapValueTypeFlag, MapValueTypeFlagShort, "", help)
	if required {
		if err := cmd.MarkFlagRequired(MapValueTypeFlag); err != nil {
			panic(err)
		}
	}
	err := cmd.RegisterFlagCompletionFunc(MapValueTypeFlag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return internal.SupportedTypeNames, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		panic(err)
	}
}

func decorateCommandWithMapKeyFlags(cmd *cobra.Command, mapKey *string, required bool, usage string) {
	cmd.Flags().StringVarP(mapKey, MapKeyFlag, MapKeyFlagShort, "", usage)
	if required {
		err := cmd.MarkFlagRequired(MapKeyFlag)
		if err != nil {
			panic(err)
		}
	}
}

func decorateCommandWithMapKeyArrayFlags(cmd *cobra.Command, mapKeys *[]string, required bool, usage string) {
	cmd.Flags().StringArrayVarP(mapKeys, MapKeyFlag, MapKeyFlagShort, []string{}, usage)
	if required {
		if err := cmd.MarkFlagRequired(MapKeyFlag); err != nil {
			panic(err)
		}
	}
}

func decorateCommandWithMapValueArrayFlags(cmd *cobra.Command, mapValues *[]string, required bool, usage string) {
	cmd.Flags().StringArrayVarP(mapValues, MapValueFlag, MapValueFlagShort, []string{}, usage)
	if required {
		if err := cmd.MarkFlagRequired(MapValueFlag); err != nil {
			panic(err)
		}
	}
}

func decorateCommandWithMapValueFileArrayFlags(cmd *cobra.Command, mapValueFiles *[]string, required bool, usage string) {
	cmd.Flags().StringArrayVarP(mapValueFiles, MapValueFileFlag, MapValueFileFlagShort, []string{}, usage)
	if required {
		if err := cmd.MarkFlagRequired(MapValueFileFlag); err != nil {
			panic(err)
		}
	}
}
