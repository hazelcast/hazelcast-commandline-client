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
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"gopkg.in/yaml.v2"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/file"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/tuiutil"
	"github.com/hazelcast/hazelcast-commandline-client/log"
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
	Styling          Styling
	Logger           Logger
	SchemaDir        string
}

type Logger struct {
	LogFile string
}

type Styling struct {
	Theme string
	tuiutil.ColorPalette
}

type GlobalFlagValues struct {
	CfgFile          string
	Cluster          string
	Token            string
	Address          string
	Verbose          bool
	NoAutocompletion bool
	NoColor          bool
	LogFile          string
	LogLevel         string
	SchemaDir        string
}

func DefaultConfig() Config {
	hz := hazelcast.Config{}
	hz.Cluster.Unisocket = true
	hz.Logger.Level = logger.ErrorLevel
	hz.Cluster.Name = DefaultClusterName
	hz.Stats.Enabled = true
	dc := Config{Hazelcast: hz}
	return dc
}

const defaultUserConfig = `hazelcast:
  cluster:
    name: {{ .Hazelcast.Cluster.Name}}
    unisocket: {{ .Hazelcast.Cluster.Unisocket}}
    network:
      # 0s is no timeout
      connectiontimeout: {{ .Hazelcast.Cluster.Network.ConnectionTimeout}}
      addresses:
      {{- range .Hazelcast.Cluster.Network.Addresses}}
        - {{ . -}}
      {{ else }}
        - localhost:5701
      {{- end }}
    cloud:
      token: "{{ .Hazelcast.Cluster.Cloud.Token}}"
      enabled: {{ .Hazelcast.Cluster.Cloud.Enabled}}
    security:
      credentials:
        username: ""
        password: ""
    discovery:
      usepublicip: false
  logger:
    level: error
ssl:
  enabled: {{ .SSL.Enabled}}
  servername: "{{ .SSL.ServerName}}"
  capath: "{{ .SSL.CAPath}}"
  certpath: "{{ .SSL.CertPath}}"
  keypath: "{{ .SSL.KeyPath}}"
  keypassword: "{{ .SSL.KeyPassword}}"
# disables auto completion on interactive mode if true
noautocompletion: false
styling:
  # builtin themes: default, no-color, solarized
  theme: "default"
logger:
  # see hazelcast.logger.level to adjust the log level of the Hazelcast Client 
  logfile: ""
`

var ValidLogLevels = []string{
	string(logger.OffLevel),
	string(logger.FatalLevel),
	string(logger.ErrorLevel),
	string(logger.WarnLevel),
	string(logger.InfoLevel),
	string(logger.DebugLevel),
	string(logger.TraceLevel),
}

func Exists() bool {
	exists, err := file.Exists(DefaultConfigPath())
	if err != nil {
		return false
	}
	return exists
}

func WriteToFile(config *Config, confPath string) error {
	t := template.Must(template.New("config").Parse(defaultUserConfig))
	var buf bytes.Buffer
	err := t.Execute(&buf, *config)
	if err != nil {
		return err
	}
	return file.CreateMissingDirsAndFileWithRWPerms(confPath, buf.Bytes())
}

func ReadAndMergeWithFlags(flags *GlobalFlagValues, c *Config) error {
	p := DefaultConfigPath()
	if err := readConfig(flags.CfgFile, c, p); err != nil {
		return err
	}
	setStyling(flags.NoColor, c)
	return mergeFlagsWithConfig(flags, c)
}

func SetupLogger(c *Config, flags *GlobalFlagValues, logOut io.Writer) (log.Logger, error) {
	var logFile string
	out := log.NopWriteCloser(logOut)
	// flag option has the precedence
	if flags.LogFile != "" {
		logFile = flags.LogFile
	} else if c.Logger.LogFile != "" {
		logFile = c.Logger.LogFile
	}
	if logFile != "" && logFile != "stderr" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return log.Logger{}, hzcerrors.NewLoggableError(err, "Can not open/create the log file on the specified path %s", flags.LogFile)
		}
		out = f
	}
	l := log.NewLogger(out)
	// set custom Logger
	c.Hazelcast.Logger.CustomLogger = log.NewClientLogger(l.Logger, c.Hazelcast.Logger.Level)
	// unset log level. Go client raises an error if both are set
	c.Hazelcast.Logger.Level = ""
	return l, nil
}

func setStyling(noColorFlag bool, c *Config) {
	if noColorFlag {
		c.Styling.Theme = tuiutil.NoColor
	}
	styling := c.Styling
	if styling.Theme != "" {
		// if not a valid theme, leave it as default
		_ = tuiutil.SetTheme(styling.Theme)
	}
	ifSetReplace := func(org *tuiutil.Color, replacement *tuiutil.Color) {
		if replacement == nil {
			return
		}
		*org = *replacement
	}
	// Override colors if specified
	theme := tuiutil.GetTheme()
	ifSetReplace(theme.HeaderBackground, styling.HeaderBackground)
	ifSetReplace(theme.Border, styling.Border)
	ifSetReplace(theme.ResultText, styling.ResultText)
	ifSetReplace(theme.HeaderForeground, styling.HeaderForeground)
	ifSetReplace(theme.Highlight, styling.Highlight)
	ifSetReplace(theme.FooterForeground, styling.FooterForeground)
	return
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
	if flags.LogLevel != "" {
		_, err := logger.WeightForLogLevel(logger.Level(flags.LogLevel))
		if err != nil {
			return hzcerrors.NewLoggableError(err, "Invalid log level (%s), should be one of %s", flags.LogLevel, ValidLogLevels)
		}
		config.Hazelcast.Logger.Level = logger.Level(flags.LogLevel)
	}
	if flags.LogFile != "" {
		config.Logger.LogFile = flags.LogFile
	}
	// must return nil err
	verboseWeight, _ := logger.WeightForLogLevel(logger.DebugLevel)
	confLevel := config.Hazelcast.Logger.Level
	confWeight, err := logger.WeightForLogLevel(confLevel)
	if err != nil {
		return hzcerrors.NewLoggableError(err, "Invalid log level (%s) on configuration file, should be one of %s", confLevel, ValidLogLevels)
	}
	if flags.Verbose && verboseWeight > confWeight {
		config.Hazelcast.Logger.Level = logger.DebugLevel
	}
	// overwrite config if flag is set
	if flags.NoAutocompletion {
		config.NoAutocompletion = true
	}
	// TODO: refactor this to a reusable function
	schemaDir := flags.SchemaDir
	if schemaDir == "" {
		schemaDir = config.SchemaDir
	}
	if schemaDir != "" {
		serialization.UpdateConfigWithRecursivePaths(&config.Hazelcast, schemaDir)
	}
	return nil
}

func readConfig(path string, config *Config, defaultConfPath string) error {
	var confBytes []byte
	var err error
	exists, err := file.Exists(path)
	if err != nil {
		return hzcerrors.NewLoggableError(err, "can not access configuration: %s", path)
	}
	if !exists {
		if path != defaultConfPath {
			// file should exist if custom path is used
			return hzcerrors.NewLoggableError(os.ErrNotExist, "configuration not found: %s", path)
		}
		*config = DefaultConfig()
		return nil
		/*
			if err = WriteToFile(config, path); err != nil {
				return hzcerrors.NewLoggableError(err, "cannot create default configuration: %s. Make sure that process has necessary permissions to write default path.\n", path)
			}

		*/
	}
	confBytes, err = os.ReadFile(path)
	if err != nil {
		return hzcerrors.NewLoggableError(err, "cannot read configuration file on %s. Make sure Configuration path is correct and process has required permission.\n", path)
	}
	if err = yaml.Unmarshal(confBytes, config); err != nil {
		return hzcerrors.NewLoggableError(err, "%s is not valid YAML", path)
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
		address = "Viridian"
	case len(c.Cluster.Network.Addresses) > 0:
		address = c.Cluster.Network.Addresses[0]
	default:
		address = DefaultClusterAddress
	}
	return address
}
