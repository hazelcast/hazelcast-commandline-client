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
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const defaultConfigPath string = ".hzc.yaml"

func DefaultConfig() *hazelcast.Config {
	config := hazelcast.NewConfig()
	config.Logger.Level = logger.ErrorLevel
	return &config
}

func MakeConfig(cmd *cobra.Command) (*hazelcast.Config, error) {
	flags := cmd.InheritedFlags()
	config := DefaultConfig()
	var confBytes []byte
	confPath, err := flags.GetString("config")
	if err != nil {
		return nil, err
	}
	if confPath != "" {
		confBytes, err = ioutil.ReadFile(confPath)
		if err != nil {
			return nil, err
		}
		fmt.Println("read by custom config file")
	} else {
		hdir, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		if err = os.Chdir(hdir); err != nil {
			return nil, err
		}
		if _, err := os.Stat(defaultConfigPath); err != nil {
			fmt.Println("default file does not exist.")
			_, err := os.Create(defaultConfigPath)
			if err != nil {
				return nil, err
			}
			config.Cluster.Unisocket = true
			out, err := yaml.Marshal(config)
			if err != nil {
				return nil, err
			}
			err = ioutil.WriteFile(defaultConfigPath, out, 0666)
			if err != nil {
				return nil, err
			}
			fmt.Println("default config file created at `~/.hzc.yaml`")
		}
		confBytes, err = ioutil.ReadFile(defaultConfigPath)
		if err != nil {
			return nil, err
		}
		fmt.Println("read by default config file at `~/.hzc.yaml`")
	}
	yaml.Unmarshal(confBytes, config)
	if err != nil {
		return nil, err
	}
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
	return config, nil
}
