package viridian

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

func (a API) DownloadClusterLogs(ctx context.Context, destDir string, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	r, err := WithRetry(ctx, &a, func() (types.Tuple2[string, func()], error) {
		u := makeUrl(fmt.Sprintf("/cluster/%s/logs", c.ID))
		path, stop, err := download(ctx, makeUrl(u), a.Token())
		if err != nil {
			return types.Tuple2[string, func()]{}, err
		}
		return types.Tuple2[string, func()]{path, stop}, nil
	})
	if err != nil {
		return err
	}
	defer r.Second()
	zipFile, err := os.Open(r.First)
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
