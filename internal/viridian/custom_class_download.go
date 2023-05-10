package viridian

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
)

type DownloadProgressPrinter struct {
	Total   uint64
	Current uint64
	Spinner clc.Spinner
}

func (pp *DownloadProgressPrinter) Write(p []byte) (int, error) {
	n := len(p)
	pp.Current += uint64(n)
	pp.Print()
	return n, nil
}

func (pp *DownloadProgressPrinter) Print() {
	pp.Spinner.SetProgress(float32(pp.Current) / float32(pp.Total))
}

func doCustomClassDownload(ctx context.Context, sp clc.Spinner, t TargetInfo, url, className, token string) error {
	fn, err := t.fileToBeCreated(className)
	if err != nil {
		return err
	}
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, makeUrl(url), nil)
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
		return fmt.Errorf("an error occurred while downloading custom class: %w", err)
	}
	if err != nil {
		return fmt.Errorf("an error occurred while downloading custom class: %w", err)
	}
	p := &DownloadProgressPrinter{Spinner: sp, Total: uint64(res.ContentLength)}
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
		exists, err := isExists(target)
		if err != nil {
			return TargetInfo{}, err
		}
		i.IsPathExists = exists
		i.IsFile, i.FileName, i.Path = isFile(target)
		if !i.IsFile {
			i.Path = target
		}
	}
	return i, nil
}

func (t TargetInfo) IsOverwrite() bool {
	return t.IsSet && t.IsFile && t.IsPathExists
}

func (t TargetInfo) fileToBeCreated(className string) (string, error) {
	var toBeCreatedFile string
	if !t.IsSet {
		toBeCreatedFile = className
	} else {
		// if directory does not exist, create it
		err := t.createDirIfNotExists()
		if err != nil {
			return "", err
		}
		if !t.IsFile { // then it is a directory
			toBeCreatedFile = t.Path + "/" + className
		} else {
			toBeCreatedFile = t.Path + "/" + t.FileName
		}
	}
	return toBeCreatedFile, nil
}

func (t TargetInfo) createDirIfNotExists() error {
	_, err := os.Stat(t.Path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(t.Path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func isExists(target string) (bool, error) {
	_, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func isFile(target string) (bool, string, string) {
	t := strings.Split(target, "/")
	e := strings.Split(t[len(t)-1], ".")

	if len(e) == 2 { // file.extension
		return true, t[len(t)-1], strings.Join(t[:len(t)-1], "/") // return the file's path and its name
	}
	return false, "", ""
}
