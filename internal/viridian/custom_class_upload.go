package viridian

import (
	"bytes"
	"context"
	"io"

	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/errors"
)

type UploadProgressReader struct {
	Reader     io.Reader
	Total      int64
	Current    int64
	EOF        bool
	SetterFunc func(progress float32)
}

func (pr *UploadProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	if err == io.EOF {
		pr.EOF = true
	}
	pr.Print()
	return n, err
}

func (pr *UploadProgressReader) Print() {
	pr.SetterFunc(float32(pr.Current) / float32(pr.Total))
	if pr.EOF {
		pr.SetterFunc(1)
	}
}

func doCustomClassUpload(ctx context.Context, progressSetter func(progress float32), url, path, token string) error {
	reqBody := &bytes.Buffer{}
	w := multipart.NewWriter(reqBody)
	p, err := w.CreateFormFile("customClassesFile", filepath.Base(path))
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	size, err := io.Copy(p, file)
	if err != nil {
		return err
	}
	w.Close()
	pr := &UploadProgressReader{Reader: reqBody, Total: size, SetterFunc: progressSetter}
	req, err := http.NewRequest(http.MethodPost, makeUrl(url), pr)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resBody := &bytes.Buffer{}
	_, err = resBody.ReadFrom(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.NewHTTPClientError(res.StatusCode, resBody.Bytes())
	}
	return nil
}
