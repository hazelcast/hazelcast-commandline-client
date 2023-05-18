package viridian

import (
	"encoding/json"
	"fmt"
)

type HTTPClientError struct {
	code    int
	text    string
	rawResp string
}

type ErrResponse struct {
	Message string `json:"message"`
}

func NewHTTPClientError(code int, body []byte) error {
	err := HTTPClientError{
		code:    code,
		rawResp: string(body),
		// it can be overwritten
		text: "an unexpected error occurred, please check logs for details",
	}
	var resp ErrResponse
	// if there is an error, resp.Message will be empty, so we can ignore it
	json.Unmarshal(body, &resp)
	// overwriting error text
	if resp.Message != "" {
		err.text = resp.Message
	}
	return err
}

func (h HTTPClientError) Error() string {
	return fmt.Sprintf("%d: %s", h.code, h.rawResp)
}

func (h HTTPClientError) Text() string {
	return h.text
}

func (h HTTPClientError) Code() int {
	return h.code
}
