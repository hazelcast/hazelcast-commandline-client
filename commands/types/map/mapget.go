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
package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/chroma/quick"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var mapGetCmd = &cobra.Command{
	Use:   "get [--name mapname | --key keyname]",
	Short: "get from map",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.TODO()
		config, err := internal.MakeConfig(cmd)
		if err != nil {
			return err
		}
		m, err := getMap(config, mapName)
		if err != nil {
			if errors.As(err, &context.DeadlineExceeded) {
				log.Fatal(internal.ErrConnectionTimeout)
			}
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
				fmt.Println(v)
				fmt.Println(value)
			}
		}
		return nil
	},
}

func init() {
	decorateCommandWithMapNameFlags(mapGetCmd)
	decorateCommandWithKeyFlags(mapGetCmd)
}
