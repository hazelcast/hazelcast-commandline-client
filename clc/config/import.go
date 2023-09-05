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

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

func ImportSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, error) {
	target = strings.TrimSpace(target)
	src = strings.TrimSpace(src)
	// check whether this an HTTP source
	path, ok, err := tryImportHTTPSource(ctx, ec, target, src)
	if err != nil {
		return "", err
	}
	// import is successful
	if ok {
		return path, nil
	}
	// import is not successful, so assume this is a zip file path and try to import from it.
	path, ok, err = tryImportViridianZipSource(ctx, ec, target, src)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("unusable source: %s", src)
	}
	return path, nil
}

func tryImportHTTPSource(ctx context.Context, ec plug.ExecContext, target, url string) (string, bool, error) {
	if !strings.HasPrefix(url, "https://") && !strings.HasSuffix(url, "http://") {
		return "", false, nil
	}
	path, err := download(ctx, ec, url)
	if err != nil {
		return "", false, err
	}
	ec.Logger().Info("Downloaded the configuration at: %s", path)
	return tryImportViridianZipSource(ctx, ec, target, path)
}

// tryImportViridianZipSource returns true if importing from a Viridian Go sample zip file is successful
func tryImportViridianZipSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, bool, error) {
	path, ok, err := CreateFromZip(ctx, ec, target, src)
	if ok {
		return path, true, nil
	}
	path, ok, err = CreateFromZipLegacy(ctx, ec, target, src)
	if err != nil {
		return "", ok, err
	}
	return path, ok, nil
}

func download(ctx context.Context, ec plug.ExecContext, url string) (string, error) {
	p, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Downloading the configuration")
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
	stop()
	return p.(string), nil
}

func CreateFromZip(ctx context.Context, ec plug.ExecContext, target, path string) (string, bool, error) {
	// TODO: refactor this function so it is not dependent on ec
	p, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Extracting configuration files")
		reader, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		// check whether this is the new config zip
		var newConfig bool
		var files []*zip.File
		for _, rf := range reader.File {
			if strings.HasSuffix(rf.Name, "/config.json") {
				newConfig = true
			}
			if !rf.FileInfo().IsDir() {
				files = append(files, rf)
			}
		}
		if !newConfig {
			return false, nil
		}
		// this is the new config zip, just extract to target
		outDir, cfgFileName, err := DirAndFile(target)
		if err != nil {
			return nil, err
		}
		if err = os.MkdirAll(outDir, 0700); err != nil {
			return nil, err
		}
		if err = copyFiles(ec, files, outDir); err != nil {
			return nil, err
		}
		return paths.Join(outDir, cfgFileName), nil
	})
	if err != nil {
		return "", false, err
	}
	stop()
	if p == false {
		return "", false, nil
	}
	return p.(string), true, nil
}

func CreateFromZipLegacy(ctx context.Context, ec plug.ExecContext, target, path string) (string, bool, error) {
	// TODO: refactor this function so it is not dependent on ec
	p, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Extracting configuration files")
		reader, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		var pyPaths []string
		var pemFiles []*zip.File
		// find .py and .pem paths
		for _, rf := range reader.File {
			if strings.HasSuffix(rf.Name, ".py") {
				pyPaths = append(pyPaths, rf.Name)
				continue
			}
			// copy only pem files
			if !strings.HasSuffix(rf.Name, ".pem") {
				continue
			}
			pemFiles = append(pemFiles, rf)
		}
		var cfgFound bool
		// find the configuration bits
		token, clusterName, pw, apiBase, cfgFound := extractConfigFields(reader, pyPaths)
		if !cfgFound {
			return nil, errors.New("python file with configuration not found")
		}
		opts := makeViridianOpts(clusterName, token, pw, apiBase)
		outDir, cfgPath, err := Create(target, opts)
		if err != nil {
			return nil, err
		}
		mopts := makeViridianOptsMap(clusterName, token, pw, apiBase)
		// ignoring the JSON path for now
		_, _, err = CreateJSON(target, mopts)
		if err != nil {
			ec.Logger().Warn("Failed creating the JSON configuration: %s", err.Error())
		}
		// copy pem files
		if err := copyFiles(ec, pemFiles, outDir); err != nil {
			return nil, err
		}
		return paths.Join(outDir, cfgPath), nil
	})
	if err != nil {
		return "", false, err
	}
	stop()
	return p.(string), true, nil
}

func makeViridianOpts(clusterName, token, password, apiBaseURL string) types.KeyValues[string, string] {
	return types.KeyValues[string, string]{
		{Key: "cluster.name", Value: clusterName},
		{Key: "cluster.discovery-token", Value: token},
		{Key: "cluster.api-base", Value: apiBaseURL},
		{Key: "ssl.ca-path", Value: "ca.pem"},
		{Key: "ssl.cert-path", Value: "cert.pem"},
		{Key: "ssl.key-path", Value: "key.pem"},
		{Key: "ssl.key-password", Value: password},
	}
}

func makeViridianOptsMap(clusterName, token, password, apiBaseURL string) map[string]any {
	cm := map[string]any{
		"name":            clusterName,
		"discovery-token": token,
		"api-base":        apiBaseURL,
	}
	ssl := map[string]any{
		"ca-path":      "ca.pem",
		"cert-path":    "cert.pem",
		"key-path":     "key.pem",
		"key-password": password,
	}
	return map[string]any{
		"cluster": cm,
		"ssl":     ssl,
	}
}

func extractConfigFields(reader *zip.ReadCloser, pyPaths []string) (token, clusterName, pw, apiBase string, cfgFound bool) {
	for _, p := range pyPaths {
		rc, err := reader.Open(p)
		if err != nil {
			continue
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			continue
		}
		text := string(b)
		token = extractViridianToken(text)
		if token == "" {
			continue
		}
		clusterName = extractClusterName(text)
		if clusterName == "" {
			continue
		}
		pw = extractKeyPassword(text)
		// it's OK if password is not found
		apiBase = extractClusterAPIBaseURL(text)
		if apiBase != "" {
			apiBase = "https://" + apiBase
		}
		// it's OK if apiBase is not found
		cfgFound = true
		break
	}
	return
}

func copyFiles(ec plug.ExecContext, files []*zip.File, outDir string) error {
	for _, rf := range files {
		_, outFn := filepath.Split(rf.Name)
		path := paths.Join(outDir, outFn)
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		rc, err := rf.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(f, rc)
		// ignoring the error here
		_ = rc.Close()
		if err != nil {
			ec.Logger().Error(err)
		}
	}
	return nil
}

func extractClusterName(text string) string {
	// extract from cluster_name="XXXX"
	const re = `cluster_name="([^"]+)"`
	return extractSimpleString(re, text)
}

func extractClusterAPIBaseURL(text string) string {
	// extract from HazelcastCloudDiscovery._CLOUD_URL_BASE = "XXXX"
	const re = `HazelcastCloudDiscovery._CLOUD_URL_BASE\s*=\s*"([^"]+)"`
	return extractSimpleString(re, text)
}

func extractViridianToken(text string) string {
	// extract from: cloud_discovery_token="XXXX",
	const re = `cloud_discovery_token="([^"]+)"`
	return extractSimpleString(re, text)
}

func extractKeyPassword(text string) string {
	// extract from: ssl_password="XXXX",
	const re = `ssl_password="([^"]+)"`
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
