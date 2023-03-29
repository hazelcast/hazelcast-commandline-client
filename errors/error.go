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
	"errors"
	"fmt"
)

var (
	ErrUserCancelled = errors.New("cancelled")
	ErrNotDecoded    = errors.New("not decoded")
	ErrNotAvailable  = errors.New("not available")
	ErrExitWithCode  = fmt.Errorf("exit with status")
)

type ExitWithStatusError struct {
	Code int
	Err  error
}

func (ve *ExitWithStatusError) Error() string {
	return fmt.Sprintf("value error: %s", ve.Err)
}

func (ve *ExitWithStatusError) Unwrap() error {
	return ve.Err
}

func (ve *ExitWithStatusError) GetCode() int {
	return ve.Code
}

func NewExitError(code int) *ExitWithStatusError {
	return &ExitWithStatusError{
		Code: code,
		Err:  ErrExitWithCode,
	}
}
