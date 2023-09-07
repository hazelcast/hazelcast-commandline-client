package config

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	ihttp "github.com/hazelcast/hazelcast-commandline-client/internal/http"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func ImportSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, error) {
	target = strings.TrimSpace(target)
	src = strings.TrimSpace(src)
	// check whether this an HTTP source
	path, err := tryImportHTTPSource(ctx, ec, target, src)
	if err != nil {
		return "", err
	}
	// import is successful
	if path != "" {
		return path, nil
	}
	// import is not successful, so assume this is a zip file path and try to import from it.
	path, err = tryImportViridianZipSource(ctx, ec, target, src)
	if err != nil {
		return "", err
	}
	if path != "" {
		return "", fmt.Errorf("unusable source: %s", src)
	}
	return path, nil
}

func tryImportHTTPSource(ctx context.Context, ec plug.ExecContext, target, url string) (string, error) {
	if !strings.HasPrefix(url, "https://") && !strings.HasSuffix(url, "http://") {
		return "", nil
	}
	path, err := download(ctx, url)
	if err != nil {
		return "", err
	}
	ec.Logger().Info("Downloaded the configuration at: %s", path)
	return tryImportViridianZipSource(ctx, ec, target, path)
}

// tryImportViridianZipSource returns true if importing from a Viridian Go sample zip file is successful
func tryImportViridianZipSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, error) {
	path, ok, err := CreateFromZip(ctx, ec, target, src)
	if err != nil {
		return "", err
	}
	if ok {
		return path, nil
	}
	return "", nil
}

func download(ctx context.Context, url string) (string, error) {
	f, err := os.CreateTemp("", "clc-download-*")
	if err != nil {
		return "", err
	}
	defer f.Close()
	client := ihttp.NewClient()
	resp, err := client.Get(ctx, url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("downloading file: %w", err)
	}
	if err != nil {
		return "", nil
	}
	return f.Name(), nil
}

func CreateFromZip(ctx context.Context, ec plug.ExecContext, target, path string) (string, bool, error) {
	// TODO: refactor this function so it is not dependent on ec
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", false, err
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
		return "", false, nil
	}
	// this is the new config zip, just extract to target
	outDir, cfgFileName, err := DirAndFile(target)
	if err != nil {
		return "", false, err
	}
	if err = os.MkdirAll(outDir, 0700); err != nil {
		return "", false, err
	}
	if err = copyFiles(ec, files, outDir); err != nil {
		return "", false, err
	}
	return paths.Join(outDir, cfgFileName), true, nil
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
