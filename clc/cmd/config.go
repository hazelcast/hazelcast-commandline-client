package cmd

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/logger"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func makeConfiguration(props plug.ReadOnlyProperties, lg *logger.Logger) (hazelcast.Config, error) {
	// if the path is not absolute, assume it is in the parent directory of the configuration
	wd := filepath.Dir(props.GetString(clc.PropertyConfig))
	var cfg hazelcast.Config
	cfg.Logger.CustomLogger = lg
	cfg.Cluster.Unisocket = true
	cfg.Stats.Enabled = true
	if ca := props.GetString(clc.PropertyClusterAddress); ca != "" {
		lg.Debugf("Cluster address: %s", ca)
		cfg.Cluster.Network.SetAddresses(ca)
	}
	if cn := props.GetString(clc.PropertyClusterName); cn != "" {
		lg.Debugf("Cluster name: %s", cn)
		cfg.Cluster.Name = cn
	}
	sd := props.GetString(clc.PropertySchemaDir)
	if sd == "" {
		sd = paths.Join(paths.Home(), "schemas")
	}
	lg.Info("Loading schemas recursively from directory: %s", sd)
	if err := serialization.UpdateSerializationConfigWithRecursivePaths(&cfg, lg, sd); err != nil {
		lg.Error(fmt.Errorf("loading serialization paths: %w", err))
	}
	var viridianEnabled bool
	if vt := props.GetString(clc.PropertyViridianToken); vt != "" {
		lg.Debugf("Viridan token: XXX")
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
		lg.Debugf("SSL server name: %s", sn)
		tc := &tls.Config{ServerName: sn}
		sc := &cfg.Cluster.Network.SSL
		sc.Enabled = true
		sc.SetTLSConfig(tc)
		if cp := props.GetString(clc.PropertySSLCAPath); cp != "" {
			cp = paths.Join(wd, cp)
			lg.Debugf("SSL CA path: %s", cp)
			if err := sc.SetCAPath(cp); err != nil {
				return cfg, err
			}
		}
		cp := props.GetString(clc.PropertySSLCertPath)
		kp := props.GetString(clc.PropertySSLKeyPath)
		kps := props.GetString(clc.PropertySSLKeyPassword)
		if cp != "" && kp != "" {
			cp = paths.Join(wd, cp)
			lg.Debugf("Certificate path: %s", cp)
			kp = paths.Join(wd, kp)
			lg.Debugf("Key path: %s", kp)
			if kps != "" {
				lg.Debugf("Key password: XXX")
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
	cfg.ClientName = makeClientName()
	return cfg, nil
}

func makeClientName() string {
	var userName string
	u, err := user.Current()
	if err != nil {
		userName = "UNKNOWN"
	} else {
		userName = u.Username
	}
	var hostName string
	host, err := os.Hostname()
	if err != nil {
		host = "UNKNOWN"
	} else {
		hostName = host
	}
	t := time.Now().Unix()
	return fmt.Sprintf("%s@%s-%d", userName, hostName, t)
}
