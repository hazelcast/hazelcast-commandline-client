//go:build base

package config

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ImportCmd struct{}

func (cm ImportCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("import [flags] [source]")
	help := "Imports configuration from an arbitrary source"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (cm ImportCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	src := ec.Args()[0]
	return cm.importSource(ctx, ec, src)
}

func (cm ImportCmd) importSource(ctx context.Context, ec plug.ExecContext, src string) error {
	src = strings.TrimSpace(src)
	ok, err := cm.tryImportViridianCurlSource(ctx, ec, src)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	ok, err = cm.tryImportViridianZipSource(ctx, ec, src)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unusable source: %s", src)
	}
	return nil
}

func (cm ImportCmd) tryImportViridianCurlSource(ctx context.Context, ec plug.ExecContext, src string) (bool, error) {
	const reCurlSource = `curl (?P<url>.*) -o hazelcast-cloud-(?P<language>[a-z]+)-sample-client-(?P<cn>[a-z-0-9-]+)-default\.zip`
	re, err := regexp.Compile(reCurlSource)
	if err != nil {
		return false, err
	}
	grps := re.FindStringSubmatch(src)
	if len(grps) != 4 {
		return false, nil
	}
	url := grps[1]
	language := grps[2]
	cn := grps[3]
	if language != "go" {
		return false, fmt.Errorf("%s sample is not usable as a configuration source, use Go sample", language)
	}
	path, err := cm.download(ctx, ec, url)
	if err != nil {
		return false, err
	}
	ec.Logger().Info("Downloaded sample to: %s", path)
	cd := paths.ResolveConfigDir(cn)
	if err := cm.makeConfigFromZip(ctx, ec, cn, path, cd); err != nil {
		return false, err
	}
	return true, nil
}

func (cm ImportCmd) tryImportViridianZipSource(ctx context.Context, ec plug.ExecContext, src string) (bool, error) {
	const reSource = `hazelcast-cloud-(?P<language>[a-z]+)-sample-client-(?P<cn>[a-z-0-9-]+)-default\.zip`
	re, err := regexp.Compile(reSource)
	if err != nil {
		return false, err
	}
	grps := re.FindStringSubmatch(src)
	if len(grps) != 3 {
		return false, nil
	}
	language := grps[1]
	cn := grps[2]
	if language != "go" {
		return false, fmt.Errorf("%s is not usable as a configuration source, use Go sample", src)
	}
	cd := paths.ResolveConfigDir(cn)
	if err := cm.makeConfigFromZip(ctx, ec, cn, src, cd); err != nil {
		return false, err
	}
	return true, nil
}

func (cm ImportCmd) download(ctx context.Context, ec plug.ExecContext, url string) (string, error) {
	p, err := ec.ExecuteBlocking(ctx, "Downloading the sample", func(ctx context.Context) (any, error) {
		f, err := os.CreateTemp("", "clc-download-*")
		if err != nil {
			return "", err
		}
		defer f.Close()
		resp, err := http.Get(url)
		defer resp.Body.Close()
		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", fmt.Errorf("downloading file: %w", err)
		}
		return f.Name(), nil
	})
	if err != nil {
		return "", nil
	}
	return p.(string), nil
}

func (cm ImportCmd) makeConfigFromZip(ctx context.Context, ec plug.ExecContext, clusterName, path, outDir string) error {
	_, err := ec.ExecuteBlocking(ctx, "Extracting files from the sample", func(ctx context.Context) (any, error) {
		if err := os.MkdirAll(outDir, 0700); err != nil {
			return nil, err
		}
		r, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		var goPaths []string
		for _, rf := range r.File {
			if strings.HasSuffix(rf.Name, ".go") {
				goPaths = append(goPaths, rf.Name)
				continue
			}
			// copy only pem files
			if !strings.HasSuffix(rf.Name, ".pem") {
				continue
			}
			_, outFn := filepath.Split(rf.Name)
			f, err := os.Create(paths.Join(outDir, outFn))
			if err != nil {
				return nil, err
			}
			rc, err := rf.Open()
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(f, rc)
			// ignoring the error here
			_ = rc.Close()
			if err != nil {
				return nil, err
			}
		}
		var cfgFound bool
		// create the config
		for _, p := range goPaths {
			rc, err := r.Open(p)
			if err != nil {
				continue
			}
			b, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				continue
			}
			text := string(b)
			token := cm.extractViridianToken(text)
			if token == "" {
				continue
			}
			pw := cm.extractKeyPassword(text)
			// it's OK if password is not found
			cfgPath := paths.ResolveConfigPath(clusterName)
			if err := cm.createConfigYAML(cfgPath, clusterName, token, pw); err != nil {
				return nil, err
			}
			cfgFound = true
			break
		}
		if !cfgFound {
			return nil, errors.New("go file with configuration not found")
		}
		return nil, nil
	})
	return err
}

func (cm ImportCmd) createConfigYAML(path, clusterName, token, password string) error {
	text := fmt.Sprintf(`
cluster:
  name: "%s"
  viridian-token: "%s"
ssl:
  ca-path: "ca.pem"
  cert-path: "cert.pem"
  key-path: "key.pem"
  key-password: "%s"
`, clusterName, token, password)
	return os.WriteFile(path, []byte(text), 0600)
}

func (cm ImportCmd) extractViridianToken(text string) string {
	// config.Cluster.Cloud.Token = "EWEKHVOOQOjMN5mXB8OngRF4YG5aOm6N2LUEOlhdC7SWpY54hm"
	const re = `config.Cluster.Cloud.Token\s+=\s+"([^"]+)"`
	return extractSimpleString(re, text)
}

func (cm ImportCmd) extractKeyPassword(text string) string {
	// err = config.Cluster.Network.SSL.AddClientCertAndEncryptedKeyPath(certFile, keyFile, "12ee6ff601a")
	const re = `config.Cluster.Network.SSL.AddClientCertAndEncryptedKeyPath\(certFile,\s+keyFile,\s+"([^"]+)"`
	return extractSimpleString(re, text)
}

func extractSimpleString(pattern, text string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	grps := re.FindStringSubmatch(text)
	if len(grps) != 2 {
		return ""
	}
	return grps[1]
}

func init() {
	Must(plug.Registry.RegisterCommand("config:import", &ImportCmd{}))
}
