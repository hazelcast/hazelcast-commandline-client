package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"gopkg.in/yaml.v2"
)

func TestContextUtil(t *testing.T) {
	ctx := context.Background()
	conf := DefaultConfig()
	conf.Hazelcast.ClientName = "test-client"
	ctx = ToContext(ctx, &conf.Hazelcast)
	from := FromContext(ctx)
	assert.Equal(t, &conf.Hazelcast, from)
}

func TestDefaultConfig(t *testing.T) {
	conf := DefaultConfig()
	assert.Equal(t, DefaultClusterName, conf.Hazelcast.Cluster.Name)
	assert.Equal(t, logger.ErrorLevel, conf.Hazelcast.Logger.Level)
	assert.Equal(t, true, conf.Hazelcast.Cluster.Unisocket)
}

func TestReadConfig(t *testing.T) {
	defaultConf := DefaultConfig()
	customConfig := *defaultConf
	customConfig.Hazelcast.ClientName = "test-client"
	emptyFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(emptyFile.Name())
	defer emptyFile.Close()
	customConfFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(customConfFile.Name())
	defer customConfFile.Close()
	b, err := yaml.Marshal(customConfig)
	assert.Nil(t, err)
	_, err = customConfFile.Write(b)
	assert.Nil(t, err)
	tests := []struct {
		name              string
		defaultConfigPath string
		path              string
		// workaround !!
		// not comparing hazelcast.Config objects since config != unmarshal(marshal(config)) because of nil map and slices
		expectedClientName string
		wantErrWithMessage error
	}{
		{
			name:               "Path: custom path, File: does not exist, Expect: error",
			path:               path.Dir(emptyFile.Name()) + "non_existing",
			wantErrWithMessage: fmt.Errorf("configuration file can not be found on configuration path %s", path.Dir(emptyFile.Name())+"non_existing"),
		},
		{
			name:               "Path: custom path, File: is empty, Expect: Default Configuration",
			path:               emptyFile.Name(),
			expectedClientName: defaultConf.Hazelcast.ClientName,
		},
		{
			name:               "Path: custom path, File: custom config, Expect: Custom Configuration",
			path:               customConfFile.Name(),
			expectedClientName: customConfig.Hazelcast.ClientName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := DefaultConfig()
			err := readConfig(tt.path, conf, tt.defaultConfigPath)
			if tt.wantErrWithMessage != nil {
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), tt.wantErrWithMessage.Error())
				return
			}
			assert.Nil(t, err)
			conf.Hazelcast.Clone()
			assert.Equal(t, conf.Hazelcast.ClientName, tt.expectedClientName)
		})
	}
}

func TestMergeFlagsWithConfig(t *testing.T) {
	tests := []struct {
		flags          PersistentFlags
		expectedConfig *Config
		wantErr        bool
	}{
		{
			// Flags: none, Expect: Default config
			expectedConfig: DefaultConfig(),
		},
		{
			flags: PersistentFlags{
				Token: "test-token",
			},
			expectedConfig: func() *Config {
				c := DefaultConfig()
				c.Hazelcast.Cluster.Cloud.Token = "test-token"
				c.Hazelcast.Cluster.Cloud.Enabled = true
				c.SSL.ServerName = "hazelcast.cloud"
				return c
			}(),
		},
		{
			flags: PersistentFlags{
				Cluster: "test-cluster",
			},
			expectedConfig: func() *Config {
				c := DefaultConfig()
				c.Hazelcast.Cluster.Name = "test-cluster"
				return c
			}(),
		},
		{
			flags: PersistentFlags{
				Address: "localhost:8904,myserver:4343",
			},
			expectedConfig: func() *Config {
				c := DefaultConfig()
				c.Hazelcast.Cluster.Network.Addresses = []string{"localhost:8904", "myserver:4343"}
				return c
			}(),
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase-%d", i+1), func(t *testing.T) {
			c := DefaultConfig()
			if err := mergeFlagsWithConfig(tt.flags, c); (err != nil) != tt.wantErr {
				t.Errorf("mergeFlagsWithConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, c, tt.expectedConfig)
		})
	}
}
