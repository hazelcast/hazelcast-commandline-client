package config_test

import (
	"bytes"
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	hzlogger "github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/logger"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func TestMakeConfiguration_Default(t *testing.T) {
	props := plug.NewProperties()
	w := clc.NopWriteCloser{bytes.NewBuffer(nil)}
	lg := MustValue(logger.New(w, hzlogger.WeightDebug))
	Must(os.Setenv("CLC_CLIENT_NAME", "my-client"))
	cfg, err := config.MakeHzConfig(props, lg)
	require.NoError(t, err)
	var target hazelcast.Config
	target.ClientName = "my-client"
	target.Cluster.Unisocket = true
	target.Stats.Enabled = true
	target.Logger.CustomLogger = lg
	target.Serialization.Compact.SetSerializers()
	require.Equal(t, target, cfg)
}

func TestMakeConfiguration_Viridian(t *testing.T) {
	props := plug.NewProperties()
	props.Set(clc.PropertyClusterDiscoveryToken, "TOKEN")
	props.Set(clc.PropertyClusterName, "pr-3066")
	/*
		// TODO: need to figure out how to specify these config options --YT
		props.Set(clc.PropertySSLCertPath, "my-cert.pem")
		props.Set(clc.PropertySSLCAPath, "my-ca.pem")
		props.Set(clc.PropertySSLKeyPath, "my-key.pem")
		props.Set(clc.PropertySSLKeyPassword, "123456")
	*/
	w := clc.NopWriteCloser{bytes.NewBuffer(nil)}
	lg := MustValue(logger.New(w, hzlogger.WeightDebug))
	// set the client name to a known value
	Must(os.Setenv("CLC_CLIENT_NAME", "my-client"))
	cfg, err := config.MakeHzConfig(props, lg)
	require.NoError(t, err)
	var target hazelcast.Config
	target.ClientName = "my-client"
	target.Cluster.Unisocket = true
	target.Cluster.Name = "pr-3066"
	target.Cluster.Cloud.Enabled = true
	target.Cluster.Cloud.Token = "TOKEN"
	target.Cluster.Network.SSL.Enabled = true
	target.Cluster.Network.SSL.SetTLSConfig(&tls.Config{ServerName: "hazelcast.cloud"})
	target.Stats.Enabled = true
	target.Logger.CustomLogger = lg
	target.Serialization.Compact.SetSerializers()
	require.Equal(t, target, cfg)
}

func TestConfigDirFile(t *testing.T) {
	// ignoring the error
	_ = os.Setenv(paths.EnvCLCHome, "/home/clc")
	defer os.Unsetenv(paths.EnvCLCHome)
	td := MustValue(os.MkdirTemp("", "clctest-*"))
	existingDir := filepath.Join(td, "mydir")
	Must(os.MkdirAll(existingDir, 0700))
	existingFile := filepath.Join(td, "mydir", "myconfig.yaml")
	Must(os.WriteFile(existingFile, []byte{}, 0700))
	testCases := []struct {
		name       string
		path       string
		targetDir  string
		targetFile string
	}{
		{
			name:       "default config",
			path:       "default-cfg",
			targetDir:  "/home/clc/configs/default-cfg",
			targetFile: "config.yaml",
		},
		{
			name:       "existing cfg dir",
			path:       existingDir,
			targetDir:  existingDir,
			targetFile: "config.yaml",
		},
		{
			name:       "existing cfg file",
			path:       existingFile,
			targetDir:  existingDir,
			targetFile: "myconfig.yaml",
		},
		{
			name:       "nonexistent dir",
			path:       "/home/me/foo",
			targetDir:  "/home/me/foo",
			targetFile: "config.yaml",
		},
		{
			name:       "nonexistent file",
			path:       "/home/me/foo/some.file",
			targetDir:  "/home/me/foo",
			targetFile: "some.file",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, f, err := config.DirAndFile(tc.path)
			assert.NoError(t, err)
			assert.Equal(t, tc.targetDir, d)
			assert.Equal(t, tc.targetFile, f)
		})
	}
}

func TestCreateYAML(t *testing.T) {
	type KV clc.KeyValue[string, string]
	testCases := []struct {
		name string
		kvs  []KV
		want string
	}{
		{
			name: "no options",
			kvs:  []KV{},
			want: "",
		},
		{
			name: "single global option",
			kvs: []KV{
				{Key: "key", Value: "value"},
			},
			want: "key: value\n",
		},
		{
			name: "two global options",
			kvs: []KV{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			want: "key1: value1\nkey2: value2\n",
		},
		{
			name: "single section",
			kvs: []KV{
				{Key: "cluster.name", Value: "pr-3814"},
				{Key: "cluster.discovery-token", Value: "TOK123123"},
			},
			want: `cluster:
  discovery-token: TOK123123
  name: pr-3814
`,
		},
		{
			name: "two sections",
			kvs: []KV{
				{Key: "cluster.name", Value: "pr-3814"},
				{Key: "ssl.ca-path", Value: "ca.pem"},
				{Key: "ssl.cert-path", Value: "cert.pem"},
			},
			want: `cluster:
  name: pr-3814
ssl:
  ca-path: ca.pem
  cert-path: cert.pem
`,
		},
		{
			name: "global with two sections",
			kvs: []KV{
				{Key: "key1", Value: "value1"},
				{Key: "cluster.name", Value: "pr-3814"},
				{Key: "ssl.ca-path", Value: "ca.pem"},
				{Key: "ssl.cert-path", Value: "cert.pem"},
				{Key: "key2", Value: "value2"},
			},
			want: `key1: value1
key2: value2
cluster:
  name: pr-3814
ssl:
  ca-path: ca.pem
  cert-path: cert.pem
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kvs := make(clc.KeyValues[string, string], len(tc.kvs))
			for i, kv := range tc.kvs {
				kvs[i] = *(*clc.KeyValue[string, string])(&kv)
			}
			s := config.CreateYAML(kvs)
			t.Logf(s)
			assert.Equalf(t, tc.want, s, "CreateYAML(%v)", tc.kvs)
		})
	}
}
