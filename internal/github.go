package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const latestReleaseURL = "https://api.github.com/repos/hazelcast/hazelcast-commandline-client/releases"

// LatestReleaseVersion returns the latest release version, except beta ones
func LatestReleaseVersion(ctx context.Context) (string, error) {
	if _, ok := ctx.Deadline(); !ok {
		var c context.CancelFunc
		ctx, c = context.WithTimeout(ctx, 3*time.Second)
		defer c()
	}
	req, err := http.NewRequest(http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data []map[string]any
	err = json.Unmarshal(respData, &data)
	if err != nil {
		return "", err
	}
	var release map[string]any
	for _, d := range data {
		prs, ok := d["prerelease"]
		if !ok {
			continue
		}
		pr, ok := prs.(bool)
		if !ok {
			continue
		}
		if !pr {
			release = d
			break
		}
	}
	if release == nil {
		return "", fmt.Errorf("no stable release")
	}
	t, ok := release["tag_name"].(string)
	if !ok {
		return "", errors.New("fetching tag_name")
	}
	return t, nil
}
