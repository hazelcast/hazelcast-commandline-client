package viridian

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"io"
	"net/http"
	"os"
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

func doCustomClassDownload(ctx context.Context, sp clc.Spinner, url, className, outputPath, token string) error {
	if outputPath != "" {
		// if the outputPath does not exist, then create it
		_, err := os.Stat(outputPath)
		if os.IsNotExist(err) {
			err = os.MkdirAll(outputPath, os.ModePerm)
			if err != nil {
				return err
			}
		}
		className = outputPath + "/" + className
	}

	f, err := os.Create(className)
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
