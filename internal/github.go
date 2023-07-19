package internal

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const latestReleaseURL = "https://api.github.com/repos/hazelcast/hazelcast-commandline-client/releases"

// LatestReleaseVersion returns the latest release version, except beta ones
func LatestReleaseVersion() (string, error) {
	resp, err := http.Get(latestReleaseURL)
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
		if d["prerelease"].(bool) == false {
			release = d
			break
		}
	}
	t, ok := release["tag_name"].(string)
	if !ok {
		return "", errors.New("fetching tag_name")
	}
	return t, nil
}
