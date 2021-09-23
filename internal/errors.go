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
package internal

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

// Error types
const (
	Network = "Network"
	Map     = "Map Operation"
)

// hzError Internal error type to differentiate known and unknown errors
type hzError struct {
	message   string
	errorType string
}

func NewHzError(errorType, message string) *hzError {
	return &hzError{errorType: errorType, message: message}
}

func (hzerr *hzError) Error() string {
	if hzerr == nil {
		return ""
	}
	return fmt.Sprintf("[%s Error]: %s", hzerr.errorType, hzerr.message)
}

func (hzerr *hzError) Type() string {
	if hzerr == nil {
		return ""
	}
	return hzerr.Type()
}

func NewHzMapError(message string) *hzError {
	return NewHzError(Map, message)
}

// Map errors
var (
	ErrMapKeyMissing                    = NewHzMapError("map key is required")
	ErrMapValueMissing                  = NewHzMapError("map value is required")
	ErrMapValueAndFileMutuallyExclusive = NewHzMapError("only one of --value and --value-file must be specified")
)

func NewHzNetworkError(message string) *hzError {
	return NewHzError(Network, message)
}

// Network errors
var (
	ErrTimeout           = NewHzNetworkError("connection timed out")
	ErrUnknownHost       = NewHzNetworkError("destination address cannot be resolved or unreachable, please make sure hazelcast cluster is up and running")
	ErrConnectionRefused = NewHzNetworkError("connection refused")
	ErrInvalidAddress    = NewHzNetworkError("invalid address")
)

func ErrorRecover() {
	obj := recover()
	if err, ok := obj.(error); ok {
		fmt.Println(err)
	}
}

func HandleNetworkError(err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ErrTimeout
		}

		/*var addrErr *net.AddrError
		if errors.As(err, &addrErr) {
			return fmt.Errorf("%v:%s", ErrInvalidAddress, addrErr.Error())
		} */
	}

	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		if netOpErr.Op == "dial" {
			return ErrUnknownHost
		} else if netOpErr.Op == "read" {
			return ErrConnectionRefused
		}
	}

	var syscallErr syscall.Errno
	//TODO these syscall errors seem platform specific, decide on what to do
	if errors.As(err, &syscallErr) && syscallErr == syscall.ECONNREFUSED {
		return ErrConnectionRefused
	}

	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		return fmt.Errorf("%v, %s", ErrInvalidAddress, addrErr.Error())
	}

	// Cannot decide on error, leave it as is, unknown
	return err
}
