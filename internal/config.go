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
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
)

const defaultConfigFilename = "config.yaml"

const (
	DefaultClusterAddress = "localhost:5701"
	DefaultClusterName    = "dev"
)

type SSLConfig struct {
	ServerName         string
	InsecureSkipVerify bool
	CAPath             string
	CertPath           string
	KeyPath            string
	KeyPassword        string
}

type Config struct {
	Hazelcast hazelcast.Config
	SSL       SSLConfig
}

// TODO: remove global Configuration

var (
	Configuration *hazelcast.Config
	CfgFile       string
	Cluster       string
	Token         string
	Address       string
)

func DefaultConfig() *Config {
	hz := hazelcast.Config{}
	hz.Cluster.Unisocket = true
	hz.Logger.Level = logger.ErrorLevel
	return &Config{Hazelcast: hz}
}

func registerConfig(config *Config, confPath string) error {
	var err error
	var out []byte
	if out, err = yaml.Marshal(config); err != nil {
		return err
	}

	filePath, _ := filepath.Split(confPath)
	if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
		return fmt.Errorf("cannot create parent directories for Configuration file: %w", err)
	}

	if err = ioutil.WriteFile(confPath, out, 0600); err != nil {
		return fmt.Errorf("writing default Configuration: %w", err)
	}
	fmt.Printf("Default Configuration file for command line client is created at `%s`\n", confPath)
	return nil
}

func validateConfig(config *Config, confPath string) error {
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

func MakeConfig() (*hazelcast.Config, error) {
	if Configuration != nil {
		return Configuration, nil
	}
	config := DefaultConfig()
	var confBytes []byte
	confPath := CfgFile
	var err error

	if confPath != DefautConfigPath() {
		confBytes, err = ioutil.ReadFile(confPath)
		if err != nil {
			fmt.Printf("Error: Cannot read Configuration file on %s. Make sure Configuration path is correct and process have sufficient permission.\n", confPath)
			return nil, fmt.Errorf("reading Configuration at %s: %w", confPath, err)
		}
	} else {
		confPath = DefautConfigPath()
		if err := validateConfig(config, confPath); err != nil {
			fmt.Printf("Error: Cannot create default Configuration file on default config path %s. Check that process has necessary permissions to write to default config path or provide a custom config path\n", confPath)
			return nil, err
		}
		if confBytes, err = ioutil.ReadFile(confPath); err != nil {
			fmt.Printf("Error: Cannot read Configuration file on default config path %s. Make sure process have sufficient permission to access Configuration path", confPath)
			return nil, fmt.Errorf("reading Configuration at %s: %w", confPath, err)
		}
	}
	if err = yaml.Unmarshal(confBytes, config); err != nil {
		fmt.Println("Error: Configuration file is not a valid yaml file, configuration read from", confPath)
		return nil, fmt.Errorf("error reading Configuration at %s: %w", confPath, err)
	}
	if Token != "" {
		config.Hazelcast.Cluster.Cloud.Token = strings.TrimSpace(Token)
		config.Hazelcast.Cluster.Cloud.Enabled = true
	}
	if err := UpdateConfigWithSSL(&config.Hazelcast, &config.SSL); err != nil {
		// TODO: Caller of MakeConfig should handle the errors or at least we should pass them to a function like ShowError, instead of printing them directly.
		// TODO: The following error does not conform to idiomatic Go, let's refactor these.
		fmt.Println(fmt.Errorf("Error configuring SSL: %W", err).Error())
		return nil, err
	}
	addrRaw := Address
	if addrRaw != "" {
		addresses := strings.Split(strings.TrimSpace(addrRaw), ",")
		config.Hazelcast.Cluster.Network.Addresses = addresses
	} else if len(config.Hazelcast.Cluster.Network.Addresses) == 0 {
		addresses := []string{DefaultClusterAddress}
		// TODO: remove this
		if config.Hazelcast.Cluster.Cloud.Enabled {
			addresses = []string{"hazelcast-cloud"}
		}
		config.Hazelcast.Cluster.Network.Addresses = addresses
	}
	if Cluster != "" {
		config.Hazelcast.Cluster.Name = strings.TrimSpace(Cluster)
	}
	if config.Hazelcast.Cluster.Cloud.Enabled {
		config.SSL.ServerName = "hazelcast.cloud"
		config.SSL.InsecureSkipVerify = false
	}
	Configuration = &config.Hazelcast
	return Configuration, nil
}

func DefautConfigPath() string {
	homeDirectoryPath, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("retrieving home directory: %w", err))
	}
	return filepath.Join(homeDirectoryPath, ".local/share/hz-cli", defaultConfigFilename)
}

func UpdateConfigWithSSL(config *hazelcast.Config, sslc *SSLConfig) error {
	if sslc.ServerName != "" && sslc.InsecureSkipVerify {
		return fmt.Errorf("SSL.ServerName and SSL.InsecureSkipVerify are mutually exclusive")
	}
	var tlsc *tls.Config
	if sslc.ServerName != "" {
		tlsc = &tls.Config{ServerName: sslc.ServerName}
	} else if sslc.InsecureSkipVerify {
		tlsc = &tls.Config{InsecureSkipVerify: true}
	}
	csslc := &config.Cluster.Network.SSL
	if tlsc != nil {
		csslc.SetTLSConfig(tlsc)
		csslc.Enabled = true
	}
	if sslc.CAPath != "" {
		if err := csslc.SetCAPath(sslc.CAPath); err != nil {
			return err
		}
	}
	if sslc.CertPath != "" || sslc.KeyPath != "" {
		if sslc.CertPath == "" {
			return fmt.Errorf("CertPath should not be blank")
		}
		if sslc.KeyPath == "" {
			return fmt.Errorf("KeyPath should not be blank")
		}
		if sslc.KeyPassword == "" {
			if err := csslc.AddClientCertAndKeyPath(sslc.CertPath, sslc.KeyPath); err != nil {
				return err
			}
		} else if err := csslc.AddClientCertAndEncryptedKeyPath(sslc.CertPath, sslc.KeyPath, sslc.KeyPassword); err != nil {
			return err
		}
	}
	return nil
}
