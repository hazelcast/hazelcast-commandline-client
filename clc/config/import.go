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
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	ihttp "github.com/hazelcast/hazelcast-commandline-client/internal/http"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func MakeImportStages(ec plug.ExecContext, target string) []stage.Stage[string] {
	stages := []stage.Stage[string]{
		{
			ProgressMsg: "Retrieving the configuration",
			SuccessMsg:  "Retrieved the configuration",
			FailureMsg:  "Failed retrieving the configuration",
			Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
				source := status.Value()
				if !strings.HasPrefix(source, "https://") || strings.HasSuffix(source, "http://") {
					if !paths.Exists(source) {
						return "", fmt.Errorf("%s does not exist", source)
					}
					return source, nil
				}
				path, err := download(ctx, source)
				if err != nil {
					return "", err
				}
				return path, nil
			},
		},
		{
			ProgressMsg: "Preparing the configuration",
			SuccessMsg:  "The configuration is ready",
			FailureMsg:  "Failed preparing the configuration",
			Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
				path := status.Value()
				path, err := CreateFromZip(ctx, target, path, ec.Logger())
				if err != nil {
					return "", err
				}
				return path, nil
			},
		},
	}
	return stages
}

/*
func ImportSource(ctx context.Context, ec plug.ExecContext, target, src string) (string, error) {
	target = strings.TrimSpace(target)
	src = strings.TrimSpace(src)
	// check whether this is an HTTP source
	path, err := tryImportHTTPSource(ctx, ec, target, src)
	if err != nil {
		return "", err
	}
	if path != "" {
		// import is successful
		return path, nil
	}
	// import is not successful, so assume this is a zip file path and try to import from it.
	path, err = tryImportViridianZipSource(ctx, target, src, ec.Logger())
	if err != nil {
		return "", err
	}
	if path != "" {
		return "", fmt.Errorf("unusable source: %s", src)
	}
	return path, nil
}

func tryImportHTTPSource(ctx context.Context, target, url string, lg log.Logger) (string, error) {
	path, err := download(ctx, url)
	if err != nil {
		return "", err
	}
	lg.Info("Downloaded the configuration at: %s", path)
	return tryImportViridianZipSource(ctx, target, path, lg)
}
*/

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

func CreateFromZip(ctx context.Context, target, path string, lg log.Logger) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	// check whether this is the new config zip
	var newConfig bool
	var files []*zip.File
	for _, rf := range reader.File {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if strings.HasSuffix(rf.Name, "/config.json") {
			newConfig = true
		}
		if !rf.FileInfo().IsDir() {
			files = append(files, rf)
		}
	}
	if !newConfig {
		return "", nil
	}
	// this is the new config zip, just extract to target
	outDir, cfgFileName, err := DirAndFile(target)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(outDir, 0700); err != nil {
		return "", err
	}
	if err = copyFiles(ctx, files, outDir, lg); err != nil {
		return "", err
	}
	return paths.Join(outDir, cfgFileName), nil
}

func copyFiles(ctx context.Context, files []*zip.File, outDir string, lg log.Logger) error {
	for _, rf := range files {
		if ctx.Err() != nil {
			return ctx.Err()
		}
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
			lg.Error(err)
		}
	}
	return nil
}
