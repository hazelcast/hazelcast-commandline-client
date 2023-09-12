package config

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"golang.org/x/exp/slices"
	"software.sslmate.com/src/go-pkcs12"

	pkcs8 "github.com/pavlo-v-chernykh/keystore-go/v4"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

const (
	envClientName   = "CLC_CLIENT_NAME"
	envClientLabels = "CLC_CLIENT_LABELS"
)

func Create(path string, opts types.KeyValues[string, string]) (dir, cfgPath string, err error) {
	return createFile(path, func(cfgPath string) (string, []byte, error) {
		text := CreateYAML(opts)
		return cfgPath, []byte(text), nil
	})
}

func CreateJSON(path string, opts map[string]any) (dir, cfgPath string, err error) {
	return createFile(path, func(cfgPath string) (string, []byte, error) {
		cfgPath = paths.ReplaceExt(cfgPath, ".json")
		b, err := json.MarshalIndent(opts, "", "  ")
		if err != nil {
			return "", nil, err
		}
		return cfgPath, b, nil
	})
}

func ConvertKeyValuesToMap(kvs types.KeyValues[string, string]) map[string]any {
	m := map[string]any{}
	for _, kv := range kvs {
		mp := m
		ps := strings.Split(kv.Key, ".")
		var i int
		var p string
		for i, p = range ps {
			if i >= len(ps)-1 {
				// this is the leaf
				break
			}
			v, ok := mp[p]
			if ok {
				// found the sub, set the map pointer
				mp = v.(map[string]any)
			} else {
				// sub doesn't exist, create it
				mm := map[string]any{}
				mp[p] = mm
				// set the map pointer
				mp = mm
			}
		}
		if p != "" {
			mp[p] = kv.Value
		}
	}
	return m
}

func createFile(path string, f func(string) (string, []byte, error)) (dir, cfgPath string, err error) {
	dir, cfgPath, err = DirAndFile(path)
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", "", err
	}
	cfgPath, b, err := f(cfgPath)
	if err != nil {
		return "", "", err
	}
	path = filepath.Join(dir, cfgPath)
	if err := os.WriteFile(path, b, 0600); err != nil {
		return "", "", err
	}
	return dir, cfgPath, nil
}

func MakeHzConfig(props plug.ReadOnlyProperties, lg log.Logger) (hazelcast.Config, error) {
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
	var viridianEnabled bool
	if vt := props.GetString(clc.PropertyClusterDiscoveryToken); vt != "" {
		lg.Debugf("Viridan token: XXX")
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
		sv := props.GetBool(clc.PropertySSLSkipVerify)
		if sv {
			lg.Debugf("Skip verify: %t", sv)
		}
		tc := &tls.Config{
			ServerName:         sn,
			InsecureSkipVerify: sv,
		}
		sc := &cfg.Cluster.Network.SSL
		sc.Enabled = true
		sc.SetTLSConfig(tc)
		if cp := props.GetString(clc.PropertySSLCAPath); cp != "" {
			cpa := strings.Split(cp, ",")
			for _, cp := range cpa {
				cp = paths.Join(wd, cp)
				lg.Debugf("SSL CA path: %s", cp)
				if err := sc.SetCAPath(cp); err != nil {
					return cfg, err
				}
			}
		}
		if ksp := props.GetString(clc.PropertySSLJavaKeyStorePath); ksp != "" {
			if kspw := props.GetString(clc.PropertySSLJavaKeyStorePassword); kspw != "" {
				if kst := props.GetString(clc.PropertySSLJavaKeyStoreType); kst != "" {
					if tsb, err := os.ReadFile(ksp); err != nil {
						return cfg, err
					} else {
						switch kst {
						case "JKS":
							ks := pkcs8.New()
							if err := ks.Load(bytes.NewReader(tsb), []byte(kspw)); err != nil {
								return cfg, err
							}
							for _, a := range ks.Aliases() {
								pke, err := ks.GetPrivateKeyEntry(a, []byte(kspw))
								if err != nil {
									return cfg, err
								}
								ca, err := x509.ParseCertificate(pke.CertificateChain[0].Content)
								if err != nil {
									return cfg, err
								}
								pk, err := x509.ParsePKCS8PrivateKey(pke.PrivateKey)
								if err != nil {
									return cfg, err
								}
								for _, cert := range pke.CertificateChain {
									if cert, err := x509.ParseCertificate(cert.Content); err != nil {
										return cfg, err
									} else if err := sc.AddClientCertAndEncryptedKey(cert, pk, ca); err != nil {
										return cfg, err
									}
								}
							}
						case "PKCS12":
							pk, ca, chain, err := pkcs12.DecodeChain(tsb, kspw)
							if err != nil {
								return cfg, err
							}
							chain = append([]*x509.Certificate{ca}, chain...)
							for _, cert := range chain {
								if err := sc.AddClientCertAndEncryptedKey(cert, pk, ca); err != nil {
									return cfg, err
								}
							}
						}
					}
				}
			}
		}
		if tsp := props.GetString(clc.PropertySSLJavaTrustStorePath); tsp != "" {
			if tspw := props.GetString(clc.PropertySSLJavaTrustStorePassword); tspw != "" {
				if tst := props.GetString(clc.PropertySSLJavaTrustStoreType); tst != "" {
					if tsb, err := os.ReadFile(tsp); err != nil {
						return cfg, err
					} else {
						switch tst {
						case "JKS":
							ks := pkcs8.New()
							if err := ks.Load(bytes.NewReader(tsb), []byte(tspw)); err != nil {
								return cfg, err
							}
							for _, a := range ks.Aliases() {
								if tce, err := ks.GetTrustedCertificateEntry(a); err != nil {
									return cfg, err
								} else if cert, err := x509.ParseCertificate(tce.Certificate.Content); err != nil {
									return cfg, err
								} else {
									sc.AddCACert(cert)
								}
							}
						case "PKCS12":
							if tsc, err := pkcs12.DecodeTrustStore(tsb, tspw); err != nil {
								return cfg, err
							} else {
								for _, cert := range tsc {
									sc.AddCACert(cert)
								}
							}
						}
					}
				}
			}
		}
		cp := props.GetString(clc.PropertySSLCertPath)
		kp := props.GetString(clc.PropertySSLKeyPath)
		cpa := strings.Split(cp, ",")
		kpa := strings.Split(kp, ",")
		kps := props.GetString(clc.PropertySSLKeyPassword)
		if len(cpa) != len(kpa) {
			panic("slices have different length")
		}
		for i := 0; i < len(cpa); i++ {
			cp = cpa[i]
			kp = kpa[i]
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
	}
	apiBase := props.GetString(clc.PropertyClusterAPIBase)
	if apiBase != "" {
		lg.Debugf("Viridan API Base: %s", apiBase)
		cfg.Cluster.Cloud.ExperimentalAPIBaseURL = apiBase
	}
	cfg.Serialization.SetIdentifiedDataSerializableFactories(serialization.SnapshotFactory{})
	cfg.Labels = makeClientLabels()
	cfg.ClientName = makeClientName()
	usr := props.GetString(clc.PropertyClusterUser)
	pass := props.GetString(clc.PropertyClusterPassword)
	if usr != "" || pass != "" {
		lg.Debugf("Cluster user: %s", usr)
		lg.Debugf("Cluster password: XXX")
		cfg.Cluster.Security.Credentials.Username = usr
		cfg.Cluster.Security.Credentials.Password = pass
	}
	return cfg, nil
}

func makeClientName() string {
	cn := os.Getenv(envClientName)
	if cn != "" {
		return cn
	}
	t := time.Now().Unix()
	return fmt.Sprintf("%s-%d", userHostName(), t)
}

func makeClientLabels() []string {
	lss, ok := os.LookupEnv(envClientLabels)
	if ok {
		return str.SplitByComma(lss, true)
	}
	return []string{"CLC", fmt.Sprintf("User:%s", userHostName())}
}

func userName() string {
	u, err := user.Current()
	if err != nil {
		return "UNKNOWN"
	}
	return u.Username
}

func hostName() string {
	host, err := os.Hostname()
	if err != nil {
		return "UNKNOWN"
	}
	return host
}

func userHostName() string {
	return fmt.Sprintf("%s@%s", userName(), hostName())
}

// DirAndFile returns the configuration directory and file separately
func DirAndFile(path string) (string, string, error) {
	path = filepath.ToSlash(path)
	// easy case, path is just a config name
	if strings.Index(path, "/") < 0 {
		return paths.ResolveConfigDir(path), paths.DefaultConfig, nil
	}
	fi, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", "", err
	}
	// if path exists, return early
	if !errors.Is(err, os.ErrNotExist) {
		if fi.IsDir() {
			return path, paths.DefaultConfig, nil
		}
		// path is a directory, split and send
		d, f := filepath.Split(path)
		return strings.TrimSuffix(d, "/"), f, err
	}
	// if path doesn't exist, check whether it's file like
	ext := filepath.Ext(path)
	if ext == "" {
		// this is probably a directory
		return path, paths.DefaultConfig, nil
	}
	// this is a file
	d, f := filepath.Split(path)
	return strings.TrimSuffix(d, "/"), f, nil
}

func CreateYAML(opts types.KeyValues[string, string]) string {
	// TODO: refactor this function to be more robust, probably using Viper
	sb := &strings.Builder{}
	copySection("", 0, sb, opts)
	return sb.String()
}

func copySection(name string, level int, sb *strings.Builder, opts types.KeyValues[string, string]) {
	slices.SortFunc(opts, func(a, b types.KeyValue[string, string]) bool {
		return a.Key < b.Key
	})
	if len(opts) == 0 {
		return
	}
	var leaves types.KeyValues[string, string]
	var sect types.KeyValues[string, string]
	sub := map[string]types.KeyValues[string, string]{}
	for _, opt := range opts {
		idx := strings.Index(opt.Key, ".")
		if idx < 0 {
			leaves = append(leaves, opt)
			continue
		}
		kh, kr := opt.Key[:idx], opt.Key[idx+1:]
		if name == "" {
			opt.Key = kr
			sub[kh] = append(sub[kh], opt)
			continue
		}
		if strings.Index(opt.Key, ".") < 0 {
			sect = append(sect, opt)
			continue
		}
		opt.Key = kr
		sub[kh] = append(sub[kh], opt)
	}
	if name != "" {
		sb.WriteString(strings.Repeat(" ", level*2))
		sb.WriteString(name)
		sb.WriteString(":\n")
		level++
	}
	for _, opt := range leaves {
		copyOpt(level, sb, opt)
	}
	for _, opt := range sect {
		copyOpt(level, sb, opt)
	}
	subSlice := make([]types.KeyValue[string, types.KeyValues[string, string]], 0, len(sub))
	for k, v := range sub {
		slices.SortFunc(v, func(a, b types.KeyValue[string, string]) bool {
			return a.Key < b.Key
		})
		subSlice = append(subSlice, types.KeyValue[string, types.KeyValues[string, string]]{
			Key:   k,
			Value: v,
		})
	}
	slices.SortFunc(subSlice, func(a, b types.KeyValue[string, types.KeyValues[string, string]]) bool {
		return a.Key < b.Key
	})
	for _, ss := range subSlice {
		copySection(ss.Key, level, sb, ss.Value)
	}
}

func copyOpt(level int, sb *strings.Builder, opt types.KeyValue[string, string]) {
	sb.WriteString(strings.Repeat(" ", level*2))
	sb.WriteString(opt.Key)
	sb.WriteString(": ")
	sb.WriteString(opt.Value)
	sb.WriteString("\n")
}

func FindAll(cd string) ([]string, error) {
	return paths.FindAll(cd, func(base string, e os.DirEntry) (ok bool) {
		return e.IsDir() && paths.Exists(paths.Join(base, e.Name(), "config.yaml"))
	})
}
