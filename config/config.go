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
package config

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"gopkg.in/yaml.v2"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
)

const defaultConfigFilename = "config.yaml"
const (
	DefaultClusterAddress = "localhost:5701"
	DefaultClusterName    = "dev"
)
const HZCConfKey = "hzc-config"

func ToContext(ctx context.Context, conf *hazelcast.Config) context.Context {
	return context.WithValue(ctx, HZCConfKey, conf)
}

func FromContext(ctx context.Context) *hazelcast.Config {
	return ctx.Value(HZCConfKey).(*hazelcast.Config)
}

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

type GlobalFlagValues struct {
	CfgFile string
	Cluster string
	Token   string
	Address string
	Verbose bool
}

func DefaultConfig() *Config {
	hz := hazelcast.Config{}
	hz.Cluster.Unisocket = true
	hz.Logger.Level = logger.ErrorLevel
	hz.Cluster.Name = DefaultClusterName
	return &Config{Hazelcast: hz}
}

func fileExists(path string) (bool, error) {
	var err error
	if _, err = os.Stat(path); err == nil {
		// conf file exists
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	// unexpected error
	return false, err
}

func writeToFile(config *Config, confPath string) error {
	var err error
	var out []byte
	if out, err = yaml.Marshal(config); err != nil {
		return err
	}
	filePath, _ := filepath.Split(confPath)
	if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
		return fmt.Errorf("can not create parent directories for file: %w", err)
	}
	if err = ioutil.WriteFile(confPath, out, 0600); err != nil {
		return fmt.Errorf("can not write to %s: %w", confPath, err)
	}
	return nil
}

func Get(flags GlobalFlagValues) (*hazelcast.Config, error) {
	c := DefaultConfig()
	p := DefaultConfigPath()
	if err := readConfig(flags.CfgFile, c, p); err != nil {
		return nil, err
	}
	if err := mergeFlagsWithConfig(flags, c); err != nil {
		return nil, err
	}
	return &c.Hazelcast, nil
}

func mergeFlagsWithConfig(flags GlobalFlagValues, config *Config) error {
	if flags.Token != "" {
		config.Hazelcast.Cluster.Cloud.Token = strings.TrimSpace(flags.Token)
		config.Hazelcast.Cluster.Cloud.Enabled = true
	}
	if err := updateConfigWithSSL(&config.Hazelcast, &config.SSL); err != nil {
		return hzcerror.NewLoggableError(err, "can not configure ssl")
	}
	addrRaw := flags.Address
	if addrRaw != "" {
		addresses := strings.Split(strings.TrimSpace(addrRaw), ",")
		config.Hazelcast.Cluster.Network.Addresses = addresses
	}
	if flags.Cluster != "" {
		config.Hazelcast.Cluster.Name = strings.TrimSpace(flags.Cluster)
	}
	if config.Hazelcast.Cluster.Cloud.Enabled {
		config.SSL.ServerName = "hazelcast.cloud"
		config.SSL.InsecureSkipVerify = false
	}
	// must return nil err
	verboseWeight, _ := logger.WeightForLogLevel(logger.DebugLevel)
	confLevel := config.Hazelcast.Logger.Level
	confWeight, err := logger.WeightForLogLevel(confLevel)
	if err != nil {
		validLogLevels := []logger.Level{logger.OffLevel, logger.FatalLevel, logger.ErrorLevel, logger.WarnLevel, logger.InfoLevel, logger.DebugLevel, logger.TraceLevel}
		return hzcerror.NewLoggableError(err, "Invalid log level (%s) on configuration file, should be one of %s", confLevel, validLogLevels)
	}
	if flags.Verbose && verboseWeight > confWeight {
		config.Hazelcast.Logger.Level = logger.DebugLevel
	}
	return nil
}

func readConfig(path string, config *Config, defaultConfPath string) error {
	isDefaultConfigPath := path == defaultConfPath
	var confBytes []byte
	var err error
	exists, err := fileExists(path)
	if err != nil {
		return hzcerror.NewLoggableError(err, "can not access configuration path %s", path)
	}
	if !exists && !isDefaultConfigPath {
		// file should exist if custom path is used
		return hzcerror.NewLoggableError(os.ErrNotExist, "configuration file can not be found on configuration path %s", path)
	}
	if !exists && isDefaultConfigPath {
		if err = writeToFile(config, path); err != nil {
			return hzcerror.NewLoggableError(err, "Cannot create configuration file on default configuration path %s. Make sure that process has necessary permissions to write default path.\n", path)
		}
	}
	confBytes, err = ioutil.ReadFile(path)
	if err != nil {
		return hzcerror.NewLoggableError(err, "cannot read configuration file on %s. Make sure Configuration path is correct and process has required permission.\n", path)
	}
	if err = yaml.Unmarshal(confBytes, config); err != nil {
		return hzcerror.NewLoggableError(err, "configuration file(%s) is not in yaml format", path)
	}
	return nil
}

func DefaultConfigPath() string {
	homeDirectoryPath, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("retrieving home directory: %w", err))
	}
	return filepath.Join(homeDirectoryPath, ".local/share/hz-cli", defaultConfigFilename)
}

func updateConfigWithSSL(config *hazelcast.Config, sslc *SSLConfig) error {
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

func GetClusterAddress(c *hazelcast.Config) string {
	var address string
	switch {
	case c.Cluster.Cloud.Enabled:
		address = "hazelcast-cloud"
	case len(c.Cluster.Network.Addresses) > 0:
		address = c.Cluster.Network.Addresses[0]
	default:
		address = DefaultClusterAddress
	}
	return address
}
