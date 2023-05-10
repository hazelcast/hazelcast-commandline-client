package viridian

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type UploadProgressReader struct {
	Reader  io.Reader
	Total   int64
	Current int64
	AtEOF   bool
	Spinner clc.Spinner
}

func (pr *UploadProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	if err == io.EOF {
		pr.AtEOF = true
	}
	pr.Print()
	return n, err
}

func (pr *UploadProgressReader) Print() {
	pr.Spinner.SetProgress(float32(pr.Current) / float32(pr.Total))
	if pr.AtEOF {
		pr.Spinner.SetProgress(1)
	}
}

func doCustomClassUpload(ctx context.Context, sp clc.Spinner, url, filePath, token string) error {
	reqBody := new(bytes.Buffer)
	w := multipart.NewWriter(reqBody)
	p, err := w.CreateFormFile("customClassesFile", filepath.Base(filePath))
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	size, err := io.Copy(p, file)
	if err != nil {
		return err
	}
	w.Close()
	req, err := http.NewRequest(http.MethodPost, makeUrl(url), &UploadProgressReader{Reader: reqBody, Total: size, Spinner: sp})
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
		log.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%d: %s", res.StatusCode, resBody.String())
	}
	return nil
}
