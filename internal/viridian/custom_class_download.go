package viridian

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

type DownloadProgressPrinter struct {
	Total      uint64
	Current    uint64
	SetterFunc func(progress float32)
}

func (pp *DownloadProgressPrinter) Write(p []byte) (int, error) {
	n := len(p)
	pp.Current += uint64(n)
	pp.Print()
	return n, nil
}

func (pp *DownloadProgressPrinter) Print() {
	p := float32(pp.Current) / float32(pp.Total)
	pp.SetterFunc(p)
}

func doCustomClassDownload(ctx context.Context, progressSetter func(progress float32), t TargetInfo, url, className, token string) error {
	fn, err := t.fileToBeCreated(className)
	if err != nil {
		return err
	}
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return NewHTTPClientError(res.StatusCode, nil)
	}
	if err != nil {
		return fmt.Errorf("downloading custom class: %w", err)
	}
	p := &DownloadProgressPrinter{SetterFunc: progressSetter, Total: uint64(res.ContentLength)}
	_, err = io.Copy(f, io.TeeReader(res.Body, p))
	if err != nil {
		f.Close()
		return err
	}
	return nil
}

type TargetInfo struct {
	IsSet        bool
	IsFile       bool
	IsPathExists bool
	FileName     string
	Path         string
}

func CreateTargetInfo(target string) (TargetInfo, error) {
	var i TargetInfo
	if target != "" {
		i.IsSet = true
		i.IsPathExists = isExists(target)
		i.checkIfFile(target)
	}
	return i, nil
}

func (t *TargetInfo) IsOverwrite() bool {
	return t.IsSet && t.IsFile && t.IsPathExists
}

func (t *TargetInfo) fileToBeCreated(artifactName string) (string, error) {
	toBeCreatedFile := artifactName
	if t.IsSet {
		// if directory does not exist, create it
		err := t.createDirIfNotExists()
		if err != nil {
			return "", err
		}
		if t.IsFile {
			toBeCreatedFile = paths.Join(t.Path, t.FileName)
		} else {
			toBeCreatedFile = paths.Join(t.Path, artifactName)
		}
	}
	return toBeCreatedFile, nil
}

func (t *TargetInfo) createDirIfNotExists() error {
	if t.Path == "" {
		return nil
	}
	return os.MkdirAll(t.Path, os.ModePerm)
}

func (t *TargetInfo) checkIfFile(target string) {
	dir, f := filepath.Split(target)
	ext := filepath.Ext(f)
	t.Path = target
	if f != "" && ext != "" {
		t.IsFile = true
		t.FileName = f
		t.Path = dir
	}
}

func isExists(target string) bool {
	_, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
