package viridian

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
)

func (a API) DownloadClusterLogs(ctx context.Context, destDir string, idOrName string) error {
	cid, err := a.findClusterID(ctx, idOrName)
	if err != nil {
		return err
	}
	err = downloadLogs(ctx, destDir, fmt.Sprintf("/cluster/%s/logs", cid), a.token)
	if err != nil {
		return err
	}
	return nil
}

func downloadLogs(ctx context.Context, destDir string, url, token string) error {
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
		return fmt.Errorf("status code is: %d", res.StatusCode)
	}
	tempZip, err := saveTempFile("logzip", res.Body)
	if err != nil {
		return fmt.Errorf("creating temporary  file: %w", err)
	}
	defer tempZip.Close()
	err = unzip(tempZip, destDir)
	if err != nil {
		return fmt.Errorf("unzipping: %w", err)
	}
	return nil
}

func saveTempFile(name string, r io.Reader) (*os.File, error) {
	tempFile, err := os.CreateTemp("", name)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(tempFile, r)
	if err != nil {
		return nil, err
	}
	return tempFile, nil
}

func unzip(zipFile *os.File, destDir string) error {
	fi, err := zipFile.Stat()
	if err != nil {
		return err
	}
	zipReader, err := zip.NewReader(zipFile, fi.Size())
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
	for _, file := range zipReader.File {
		r, err := file.Open()
		if err != nil {
			return err
		}
		filePath := path.Join(destDir, file.Name)
		err = saveFile(filePath, file.FileInfo(), r)
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
