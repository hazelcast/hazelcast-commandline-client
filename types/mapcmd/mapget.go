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
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func NewGet() *cobra.Command {
	var mapName, mapKey string
	cmd := &cobra.Command{
		Use:   "get [--name mapname | --key keyname]",
		Short: "get from map",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*3)
			defer cancel()
			conf := cmd.Context().Value(config.HZCConfKey).(*hazelcast.Config)
			m, err := getMap(ctx, conf, mapName)
			if err != nil {
				return err
			}
			value, err := m.Get(ctx, mapKey)
			if err != nil {
				isCloudCluster := conf.Cluster.Cloud.Enabled
				if networkErrMsg, handled := internal.TranslateNetworkError(err, isCloudCluster); handled {
					return hzcerror.NewLoggableError(err, networkErrMsg)
				}
				return hzcerror.NewLoggableError(err, "Cannot get value for key %s from map %s", mapKey, mapName)
			}
			if value == nil {
				cmd.Println("There is no value corresponding to the provided key")
				return nil
			}
			switch v := value.(type) {
			case serialization.JSON:
				if err := quick.Highlight(cmd.OutOrStdout(), fmt.Sprintln(v.String()),
					"json", "terminal", "tango"); err != nil {
					cmd.Println(v.String())
				}
			default:
				cmd.Println(value)
			}
			return nil
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName)
	decorateCommandWithKeyFlags(cmd, &mapKey)
	return cmd
}
