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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
)

const defaultConfigFilename = "config.yaml"

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

	filePath, _ := filepath.Split(confPath)
	if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
		return fmt.Errorf("cannot create parent directories for configuration file: %w", err)
	}

	if err = ioutil.WriteFile(confPath, out, 0600); err != nil {
		return fmt.Errorf("writing default configuration: %w", err)
	}
	fmt.Printf("default config file is created at `%s`\n", confPath)
	return nil
}

func validateConfig(config *hazelcast.Config, confPath string) error {
	if _, err := os.Stat(confPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err = registerConfig(config, confPath); err != nil {
			return err
		}
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
			fmt.Printf("Error: Cannot read configuration file on %s. Please make sure configuration path is correct and process have sufficient permission.\n", confPath)
			return nil, fmt.Errorf("reading configuration at %s: %w", confPath, err)
		}
	} else {
		confPath = DefautConfigPath()
		if err := validateConfig(config, confPath); err != nil {
			fmt.Printf("Error: Cannot create default configuration file on default config path %s. Please check that process has necessary permissions to write to default config path or provide a custom config path\n", confPath)
			return nil, err
		}
		if confBytes, err = ioutil.ReadFile(confPath); err != nil {
			fmt.Printf("Error: Cannot read configuration file on default config path %s. Please make sure process have sufficient permission to access configuration path", confPath)
			return nil, fmt.Errorf("reading configuration at %s: %w", confPath, err)
		}
	}
	if err = yaml.Unmarshal(confBytes, config); err != nil {
		fmt.Println("Error: Configuration file is not a valid yaml file")
		return nil, fmt.Errorf("error reading configuration at %s: %w", confPath, err)
	}
	token, err := flags.GetString("cloud-token")
	if err != nil {
		fmt.Println("Error: Invalid value for --cloud-token")
		return nil, err
	}
	if token != "" {
		config.Cluster.Cloud.Token = strings.TrimSpace(token)
		config.Cluster.Cloud.Enabled = true
	}
	addrRaw, err := flags.GetString("address")
	if err != nil {
		fmt.Println("Error: Invalid value for --address")
		return nil, err
	}
	if addrRaw != "" {
		addresses := strings.Split(strings.TrimSpace(addrRaw), ",")
		config.Cluster.Network.Addresses = addresses
	} else {
		addresses := []string{"localhost:5701"}
		config.Cluster.Network.Addresses = addresses
	}
	cluster, err := flags.GetString("cluster-name")
	if err != nil {
		fmt.Println("Error: Invalid value for --cluster-name")
		return nil, err
	}
	if cluster != "" {
		config.Cluster.Name = strings.TrimSpace(cluster)
	} else {
		config.Cluster.Name = "dev"
	}
	return config, nil
}

func DefautConfigPath() string {
	homeDirectoryPath, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("retrieving home directory: %w", err))
	}
	return filepath.Join(homeDirectoryPath, ".local/share/hz-cli", defaultConfigFilename)
}
