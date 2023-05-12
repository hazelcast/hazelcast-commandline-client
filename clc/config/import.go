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
)

func ImportSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, error) {
	target = strings.TrimSpace(target)
	src = strings.TrimSpace(src)
	// first assume the passed string is a CURL command line, and try to import it.
	path, ok, err := tryImportViridianCurlSource(ctx, ec, target, src)
	if err != nil {
		return "", err
	}
	// import is successful
	if ok {
		return path, nil
	}
	// import is not successful, check whether this an HTTP source
	path, ok, err = tryImportHTTPSource(ctx, ec, target, src)
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

// tryImportViridianCurlSource returns true if importing from a Viridian CURL command line is successful
func tryImportViridianCurlSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, bool, error) {
	const reCurlSource = `curl (?P<url>[^\s]+)\s+`
	re, err := regexp.Compile(reCurlSource)
	if err != nil {
		return "", false, err
	}
	grps := re.FindStringSubmatch(src)
	if len(grps) < 2 {
		return "", false, nil
	}
	url := grps[1]
	return tryImportHTTPSource(ctx, ec, target, url)
}

func tryImportHTTPSource(ctx context.Context, ec plug.ExecContext, target, url string) (string, bool, error) {
	if !strings.HasPrefix(url, "https://") && !strings.HasSuffix(url, "http://") {
		return "", false, nil
	}
	path, err := download(ctx, ec, url)
	if err != nil {
		return "", false, err
	}
	ec.Logger().Info("Downloaded sample to: %s", path)
	path, err = CreateFromZip(ctx, ec, target, path)
	if err != nil {
		return "", false, err
	}
	return path, true, nil

}

// tryImportViridianZipSource returns true if importing from a Viridian Go sample zip file is successful
func tryImportViridianZipSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, bool, error) {
	const reSource = `hazelcast-cloud-(?P<language>[a-z]+)-sample-client-(?P<cn>[a-zA-Z0-9_-]+)-default\.zip`
	re, err := regexp.Compile(reSource)
	if err != nil {
		return "", false, err
	}
	grps := re.FindStringSubmatch(src)
	if len(grps) != 3 {
		return "", false, nil
	}
	language := grps[1]
	if language != "go" {
		return "", false, fmt.Errorf("%s is not usable as a configuration source, use Go sample", src)
	}
	path, err := CreateFromZip(ctx, ec, target, src)
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}

func download(ctx context.Context, ec plug.ExecContext, url string) (string, error) {
	p, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Downloading the sample")
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

func CreateFromZip(ctx context.Context, ec plug.ExecContext, target, path string) (string, error) {
	p, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Extracting files from the sample")
		reader, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		var goPaths []string
		var pemFiles []*zip.File
		// find .go and .pem paths
		for _, rf := range reader.File {
			if strings.HasSuffix(rf.Name, ".go") {
				goPaths = append(goPaths, rf.Name)
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
		token, clusterName, pw, cfgFound := extractConfigFields(reader, goPaths)
		if !cfgFound {
			return nil, errors.New("go file with configuration not found")
		}
		opts := makeViridianOpts(clusterName, token, pw)
		outDir, cfgPath, err := Create(target, opts)
		if err != nil {
			return nil, err
		}
		// copy pem files
		if err := copyFiles(ec, pemFiles, outDir); err != nil {
			return nil, err
		}
		return cfgPath, nil
	})
	if err != nil {
		return "", err
	}
	stop()
	return p.(string), nil
}

func makeViridianOpts(clusterName, token, password string) clc.KeyValues[string, string] {
	return clc.KeyValues[string, string]{
		{Key: "cluster.name", Value: clusterName},
		{Key: "cluster.discovery-token", Value: token},
		{Key: "ssl.ca-path", Value: "ca.pem"},
		{Key: "ssl.cert-path", Value: "cert.pem"},
		{Key: "ssl.key-path", Value: "key.pem"},
		{Key: "ssl.key-password", Value: password},
	}
}

func extractConfigFields(reader *zip.ReadCloser, goPaths []string) (token, clusterName, pw string, cfgFound bool) {
	for _, p := range goPaths {
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
		cfgFound = true
		break
	}
	return
}

func copyFiles(ec plug.ExecContext, files []*zip.File, outDir string) error {
	for _, rf := range files {
		_, outFn := filepath.Split(rf.Name)
		f, err := os.Create(paths.Join(outDir, outFn))
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
	// extract from config.Cluster.Name = "pr-3814"
	const re = `config.Cluster.Name\s+=\s+"([^"]+)"`
	return extractSimpleString(re, text)

}

func extractViridianToken(text string) string {
	// extract from: config.Cluster.Cloud.Token = "EWEKHVOOQOjMN5mXB8OngRF4YG5aOm6N2LUEOlhdC7SWpY54hm"
	const re = `config.Cluster.Cloud.Token\s+=\s+"([^"]+)"`
	return extractSimpleString(re, text)
}

func extractKeyPassword(text string) string {
	// extract from: err = config.Cluster.Network.SSL.AddClientCertAndEncryptedKeyPath(certFile, keyFile, "12ee6ff601a")
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
