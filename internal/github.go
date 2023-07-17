package internal

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const latestRelease = "https://api.github.com/repos/hazelcast/hazelcast-commandline-client/releases/latest"

func LatestReleaseVersion() (string, error) {
	resp, err := http.Get(latestRelease)
	if err != nil {
		return "", err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data map[string]any
	err = json.Unmarshal(respData, &data)
	if err != nil {
		return "", err
	}
	if data["tag_name"] == nil {
		return "", errors.New("fetching tag_name")
	}
	return data["tag_name"].(string), nil
}
