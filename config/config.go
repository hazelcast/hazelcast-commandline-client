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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"gopkg.in/yaml.v2"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/file"
)

const defaultConfigFilename = "config.yaml"
const (
	DefaultClusterAddress  = "localhost:5701"
	DefaultClusterName     = "dev"
	DefaultCloudServerName = "hazelcast.cloud"
)

type SSLConfig struct {
	Enabled            bool
	ServerName         string
	InsecureSkipVerify bool
	CAPath             string
	CertPath           string
	KeyPath            string
	KeyPassword        string
}

type Config struct {
	Hazelcast        hazelcast.Config
	SSL              SSLConfig
	NoAutocompletion bool
}

type GlobalFlagValues struct {
	CfgFile          string
	Cluster          string
	Token            string
	Address          string
	Verbose          bool
	NoAutocompletion bool
}

func DefaultConfig() *Config {
	hz := hazelcast.Config{}
	hz.Cluster.Unisocket = true
	hz.Logger.Level = logger.ErrorLevel
	hz.Cluster.Name = DefaultClusterName
	hz.Stats.Enabled = true
	return &Config{Hazelcast: hz}
}

const defaultUserConfig = `hazelcast:
  clientname: ""
  logger:
    level: error
  cluster:
    security:
      credentials:
        username: ""
        password: ""
    name: dev
    cloud:
      token: ""
      enabled: false
    discovery:
      usepublicip: false
    unisocket: true
  network:
    addresses:
      - "localhost:5701"
    # 0s means infinite timeout (no timeout)
    connectiontimeout: 0s
ssl:
  enabled: false
  servername: ""
  capath: ""
  certpath: ""
  keypath: ""
  keypassword: ""
# disables auto completion on interactive mode
noautocompletion: false
`

func writeToFile(config string, confPath string) error {
	return file.CreateMissingDirsAndFileWithRWPerms(confPath, []byte(config))
}

func ReadAndMergeWithFlags(flags *GlobalFlagValues, c *Config) error {
	p := DefaultConfigPath()
	if err := readConfig(flags.CfgFile, c, p); err != nil {
		return err
	}
	if err := mergeFlagsWithConfig(flags, c); err != nil {
		return err
	}
	return nil
}

func mergeFlagsWithConfig(flags *GlobalFlagValues, config *Config) error {
	if flags.Token != "" {
		config.Hazelcast.Cluster.Cloud.Token = strings.TrimSpace(flags.Token)
		config.Hazelcast.Cluster.Cloud.Enabled = true
	}
	if err := updateConfigWithSSL(&config.Hazelcast, &config.SSL); err != nil {
		return hzcerrors.NewLoggableError(err, "can not configure ssl")
	}
	addrRaw := flags.Address
	if addrRaw != "" {
		addresses := strings.Split(strings.TrimSpace(addrRaw), ",")
		config.Hazelcast.Cluster.Network.Addresses = addresses
	}
	if flags.Cluster != "" {
		config.Hazelcast.Cluster.Name = strings.TrimSpace(flags.Cluster)
	}
	// must return nil err
	verboseWeight, _ := logger.WeightForLogLevel(logger.DebugLevel)
	confLevel := config.Hazelcast.Logger.Level
	confWeight, err := logger.WeightForLogLevel(confLevel)
	if err != nil {
		validLogLevels := []logger.Level{logger.OffLevel, logger.FatalLevel, logger.ErrorLevel, logger.WarnLevel, logger.InfoLevel, logger.DebugLevel, logger.TraceLevel}
		return hzcerrors.NewLoggableError(err, "Invalid log level (%s) on configuration file, should be one of %s", confLevel, validLogLevels)
	}
	if flags.Verbose && verboseWeight > confWeight {
		config.Hazelcast.Logger.Level = logger.DebugLevel
	}
	// overwrite config if flag is set
	if flags.NoAutocompletion {
		config.NoAutocompletion = true
	}
	return nil
}

func readConfig(path string, config *Config, defaultConfPath string) error {
	isDefaultConfigPath := path == defaultConfPath
	var confBytes []byte
	var err error
	exists, err := file.Exists(path)
	if err != nil {
		return hzcerrors.NewLoggableError(err, "can not access configuration path %s", path)
	}
	if !exists && !isDefaultConfigPath {
		// file should exist if custom path is used
		return hzcerrors.NewLoggableError(os.ErrNotExist, "configuration file can not be found on configuration path %s", path)
	}
	if !exists && isDefaultConfigPath {
		if err = writeToFile(defaultUserConfig, path); err != nil {
			return hzcerrors.NewLoggableError(err, "Cannot create configuration file on default configuration path %s. Make sure that process has necessary permissions to write default path.\n", path)
		}
	}
	confBytes, err = ioutil.ReadFile(path)
	if err != nil {
		return hzcerrors.NewLoggableError(err, "cannot read configuration file on %s. Make sure Configuration path is correct and process has required permission.\n", path)
	}
	if err = yaml.Unmarshal(confBytes, config); err != nil {
		return hzcerrors.NewLoggableError(err, "configuration file(%s) is not in yaml format", path)
	}
	return nil
}

func DefaultConfigPath() string {
	return filepath.Join(file.HZCHomePath(), defaultConfigFilename)
}

func updateConfigWithSSL(config *hazelcast.Config, sslc *SSLConfig) error {
	if !sslc.Enabled {
		// SSL configuration is not set, skip
		return nil
	}
	if config.Cluster.Cloud.Enabled && sslc.ServerName == "" {
		sslc.ServerName = DefaultCloudServerName
	}
	csslc := &config.Cluster.Network.SSL
	csslc.SetTLSConfig(&tls.Config{ServerName: sslc.ServerName, InsecureSkipVerify: sslc.InsecureSkipVerify})
	csslc.Enabled = true
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
