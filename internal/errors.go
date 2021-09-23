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
	"net/url"
	"strings"
	"syscall"
)

// Map errors
var (
	ErrMapKeyMissing                    = errors.New("map key is required")
	ErrMapValueMissing                  = errors.New("map value is required")
	ErrMapValueAndFileMutuallyExclusive = errors.New("only one of --value and --value-file must be specified")
)

// Cluster errors
var (
	ErrRestAPIDisabled = errors.New("Cannot access Hazelcast REST API. Is it enabled? Check this link to find out more: " + RESTApiDocs)
)

// Network errors
var (
	ErrUnknownHost       = errors.New("destination address cannot be resolved or unreachable, please make sure hazelcast cluster is up and running")
	ErrConnectionTimeout = errors.New("Can not connect to Hazelcast Cluster. Please make sure Hazelcast cluster is up, reachable and running. Check this link to create a IMDG cluster: " + QuickStartDocs)
	ErrConnectionRefused = errors.New("connection refused")
	ErrInvalidAddress    = errors.New("invalid address")
)

func ErrorRecover() {
	obj := recover()
	if err, ok := obj.(error); ok {
		fmt.Println(err)
	}
}

func HandleClusterError(err error) (error, bool) {
	var urlError *url.Error
	if errors.As(err, &urlError) && strings.Contains(urlError.Error(), "EOF") {
		return ErrRestAPIDisabled, true
	}
	return err, false
}

func HandleNetworkError(err error) (error, bool) {
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ErrConnectionTimeout, true
		}

		var addrErr *net.AddrError
		if errors.As(err, &addrErr) {
			return fmt.Errorf("%v:%s", ErrInvalidAddress, addrErr.Error()), true
		}
	}

	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		if netOpErr.Op == "dial" {
			return ErrUnknownHost, true
		} else if netOpErr.Op == "read" {
			return ErrConnectionRefused, true
		}
	}

	var syscallErr syscall.Errno
	//TODO these syscall errors seem platform specific, decide on what to do
	if errors.As(err, &syscallErr) && syscallErr == syscall.ECONNREFUSED {
		return ErrConnectionRefused, true
	}

	// Cannot decide on error, leave it as is, unknown
	return err, false
}
