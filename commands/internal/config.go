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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const defaultConfigFilename string = ".hzc.yaml"

func DefaultConfig() *hazelcast.Config {
	config := hazelcast.NewConfig()
	config.Cluster.Unisocket = true
	config.Logger.Level = logger.ErrorLevel
	return &config
}

func registerConfig(config *hazelcast.Config, confPath string) error {
	var err error
	var out []byte
	if out, err = yaml.Marshal(config); err != nil {
		return err
	}
	if err = ioutil.WriteFile(confPath, out, 0666); err != nil {
		return fmt.Errorf("writing default configuration: %w", err)
	}
	fmt.Println("default config file is created at `~/.hzc.yaml`")
	return nil
}

func validateConfig(config *hazelcast.Config, confPath string) error {
	if _, err := os.Stat(confPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		registerConfig(config, confPath)
	}
	return nil
}

func MakeConfig(cmd *cobra.Command) (*hazelcast.Config, error) {
	flags := cmd.InheritedFlags()
	config := DefaultConfig()
	var confBytes []byte
	var confPath string
	var err error
	confPath, err = flags.GetString("config")
	if err != nil {
		return nil, err
	}
	if confPath != "" {
		confBytes, err = ioutil.ReadFile(confPath)
		if err != nil {
			return nil, fmt.Errorf("reading configuration at %s: %w", confPath, err)
		}
	} else {
		var hdir string
		if hdir, err = homedir.Dir(); err != nil {
			return nil, err
		}
		confPath = filepath.Join(hdir, defaultConfigFilename)
		if err := validateConfig(config, confPath); err != nil {
			return nil, err
		}
		if confBytes, err = ioutil.ReadFile(confPath); err != nil {
			return nil, fmt.Errorf("reading configuration at %s: %w", confPath, err)
		}
	}
	if err = yaml.Unmarshal(confBytes, config); err != nil {
		return nil, fmt.Errorf("error reading configuration at %s: %w", confPath, err)
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
