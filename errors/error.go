/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package errors

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrUserCancelled   = errors.New("cancelled")
	ErrNotDecoded      = errors.New("not decoded")
	ErrNotAvailable    = errors.New("not available")
	ErrNoClusterConfig = errors.New("no configuration was specified")
)

type WrappedError struct {
	Err error
}

func (w WrappedError) Unwrap() error {
	return w.Err
}

func (w WrappedError) Error() string {
	return w.Err.Error()
}

type HTTPError interface {
	Text() string
	Code() int
}

type HTTPClientError struct {
	code    int
	text    string
	rawResp string
}

func NewHTTPClientError(code int, body []byte) error {
	err := HTTPClientError{
		code:    code,
		rawResp: string(body),
		// it can be overwritten
		text: "an unexpected error occurred, please check logs for details",
	}
	type ErrResp struct {
		Message string `json:"message"`
	}
	var resp ErrResp
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
