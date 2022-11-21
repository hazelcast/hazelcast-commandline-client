package cmd

import (
	"bytes"
	"crypto/tls"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	hzlogger "github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/logger"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func TestMakeConfiguration_Default(t *testing.T) {
	props := plug.NewProperties()
	w := nopWriteCloser{bytes.NewBuffer(nil)}
	lg := MustValue(logger.New(w, hzlogger.WeightDebug))
	cfg, err := makeConfiguration(props, lg)
	require.NoError(t, err)
	var target hazelcast.Config
	target.Cluster.Unisocket = true
	target.Stats.Enabled = true
	target.Logger.CustomLogger = lg
	target.Serialization.Compact.SetSerializers()
	require.Equal(t, target, cfg)
}

func TestMakeConfiguration_Viridian(t *testing.T) {
	props := plug.NewProperties()
	props.Set(clc.PropertyViridianToken, "TOKEN")
	props.Set(clc.PropertyClusterName, "pr-3066")
	/*
		// TODO: need to figure out how to specify these config options --YT
		props.Set(clc.PropertySSLCertPath, "my-cert.pem")
		props.Set(clc.PropertySSLCAPath, "my-ca.pem")
		props.Set(clc.PropertySSLKeyPath, "my-key.pem")
		props.Set(clc.PropertySSLKeyPassword, "123456")
	*/
	w := nopWriteCloser{bytes.NewBuffer(nil)}
	lg := MustValue(logger.New(w, hzlogger.WeightDebug))
	cfg, err := makeConfiguration(props, lg)
	require.NoError(t, err)
	var target hazelcast.Config
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
