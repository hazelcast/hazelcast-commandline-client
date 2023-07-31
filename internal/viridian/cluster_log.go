package viridian

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

func (a API) DownloadClusterLogs(ctx context.Context, destDir string, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	zipPath, stop, err := download(ctx, makeUrl(fmt.Sprintf("/cluster/%s/logs", c.ID)), a.token)
	if err != nil {
		return fmt.Errorf("downloading cluster logs: %w", err)
	}
	defer stop()
	zipFile, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	st, err := zipFile.Stat()
	if err != nil || st.Size() == 0 {
		return fmt.Errorf("logs are not available yet, retry later")
	}
	err = unzip(zipFile, destDir)
	if err != nil {
		return err
	}
	return nil
}

func unzip(zipFile *os.File, destDir string) error {
	fi, err := zipFile.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(zipFile, fi.Size())
	if err != nil {
		return err
	}

	if destDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		destDir = dir
	}

	err = os.MkdirAll(destDir, 0700)
	if err != nil {
		return err
	}
	for _, file := range zr.File {
		err = func() error {
			r, err := file.Open()
			if err != nil {
				return err
			}
			defer r.Close()
			filePath := paths.Join(destDir, file.Name)
			err = saveFile(filePath, file.FileInfo(), r)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func saveFile(name string, info fs.FileInfo, src io.Reader) error {
	if info.IsDir() {
		return os.MkdirAll(name, info.Mode())
	}
	dst, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}
