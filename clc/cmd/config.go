package cmd

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/logger"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func makeConfiguration(props plug.ReadOnlyProperties, lg *logger.Logger) (hazelcast.Config, error) {
	// if the path is not absolute, assume it is in the parent directory of the configuration
	wd := filepath.Dir(props.GetString(clc.PropertyConfigPath))
	var cfg hazelcast.Config
	cfg.Cluster.Unisocket = true
	cfg.Stats.Enabled = true
	if ca := props.GetString(clc.PropertyClusterAddress); ca != "" {
		lg.Debugf("Set the cluster address: %s", ca)
		cfg.Cluster.Network.SetAddresses(ca)
	}
	if cn := props.GetString(clc.PropertyClusterName); cn != "" {
		lg.Debugf("Set the cluster name: %s", cn)
		cfg.Cluster.Name = cn
	}
	cfg.Logger.CustomLogger = lg
	sd := props.GetString(clc.PropertySchemaDir)
	if sd == "" {
		sd = paths.Join(paths.HomeDir(), "schemas")
	}
	lg.Info("Loading schemas recursively from directory: %s", sd)
	if err := serialization.UpdateSerializationConfigWithRecursivePaths(&cfg, lg, sd); err != nil {
		lg.Error(fmt.Errorf("loading serialization paths: %w", err))
	}
	var viridianEnabled bool
	if vt := props.GetString(clc.PropertyViridianToken); vt != "" {
		lg.Debugf("Set the Viridan token: XXX")
		if err := os.Setenv(envHzCloudCoordinatorBaseURL, viridianCoordinatorURL); err != nil {
			return cfg, fmt.Errorf("setting coordinator URL")
		}
		cfg.Cluster.Cloud.Enabled = true
		cfg.Cluster.Cloud.Token = vt
		viridianEnabled = true
	}
	if props.GetBool(clc.PropertySSLEnabled) || viridianEnabled {
		sn := "hazelcast.cloud"
		if !viridianEnabled {
			sn = props.GetString(clc.PropertySSLServerName)
		}
		lg.Debugf("Using SSL server name: %s", sn)
		tc := &tls.Config{ServerName: sn}
		sc := &cfg.Cluster.Network.SSL
		sc.Enabled = true
		sc.SetTLSConfig(tc)
		if cp := props.GetString(clc.PropertySSLCAPath); cp != "" {
			cp = paths.Join(wd, cp)
			lg.Debugf("Using SSL CA path: %s", cp)
			if err := sc.SetCAPath(cp); err != nil {
				return cfg, err
			}
		}
		cp := props.GetString(clc.PropertySSLCertPath)
		kp := props.GetString(clc.PropertySSLKeyPath)
		kps := props.GetString(clc.PropertySSLKeyPassword)
		if cp != "" && kp != "" {
			cp = paths.Join(wd, cp)
			lg.Debugf("Using certificate path: %s", cp)
			kp = paths.Join(wd, kp)
			lg.Debugf("Using key path: %s", kp)
			if kps != "" {
				lg.Debugf("Using key password: XXX")
				if err := sc.AddClientCertAndEncryptedKeyPath(cp, kp, kps); err != nil {
					return cfg, err
				}
			} else {
				if err := sc.AddClientCertAndKeyPath(cp, kp); err != nil {
					return cfg, err
				}
			}
		}
	}
	return cfg, nil
}
