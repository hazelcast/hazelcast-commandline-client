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
package internal

import (
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/spf13/cobra"
)

func DefaultConfig() *hazelcast.Config {
	config := hazelcast.NewConfig()
	config.Logger.Level = logger.ErrorLevel
	return &config
}

func MakeConfig(cmd *cobra.Command) (*hazelcast.Config, error) {
	flags := cmd.InheritedFlags()
	config := DefaultConfig()
	token, err := flags.GetString("cloud-token")
	if err != nil {
		return nil, err
	}
	if token != "" {
		config.Cluster.Cloud.Token = strings.TrimSpace(token)
		config.Cluster.Cloud.Enabled = true
	}
	addrRaw, err := flags.GetString("address")
	if err != nil {
		return nil, err
	}
	if addrRaw != "" {
		addresses := strings.Split(strings.TrimSpace(addrRaw), ",")
		config.Cluster.Network.Addresses = addresses
	}
	cluster, err := flags.GetString("cluster-name")
	if err != nil {
		return nil, err
	}
	config.Cluster.Name = strings.TrimSpace(cluster)
	config.Cluster.Unisocket = true
	return config, nil
}
