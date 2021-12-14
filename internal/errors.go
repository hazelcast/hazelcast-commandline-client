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
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"syscall"

	"github.com/hazelcast/hazelcast-go-client/hzerrors"
)

const (
	restOrClusterWriteEnabledMessage = `Cannot access Hazelcast REST endpoint.
- Is REST API enabled?
REST service is disabled in the configuration by default. It enables you to manage cluster with REST-API calls.
You should enable it on your CLUSTER MEMBERS to use the cluster commands.
Check this link to find out more: https://docs.hazelcast.com/hazelcast/latest/maintain-cluster/rest-api#enabling-rest-api

- If yes, is CLUSTER_WRITE endpoint group enabled?
Endpoints of Hazelcast REST API are grouped for fine-grained authorization. Commands such as "cluster change-state" that manipulates cluster state, must be enabled explicitly.
Check this link to find out more: https://docs.hazelcast.com/hazelcast/latest/maintain-cluster/rest-api#using-rest-endpoint-groups`
	restAPIEnabledMessage = `Cannot access Hazelcast REST API.
REST service is disabled in the configuration by default. It enables you to manage cluster with REST-API calls.
You should enable it on your CLUSTER MEMBERS to use the cluster commands.
Check this link to find out more: https://docs.hazelcast.com/hazelcast/latest/maintain-cluster/rest-api#enabling-rest-api`
)

func ErrorRecover() {
	obj := recover()
	if err, ok := obj.(error); ok {
		fmt.Println(err)
	}
}

func TranslateError(err error, isCloudCluster bool, op ...string) (string, bool) {
	if len(op) == 1 {
		if msg, handled := TranslateClusterError(err, op[0]); handled {
			return msg, true
		}
	}
	return TranslateNetworkError(err, isCloudCluster)
}

func TranslateClusterError(err error, operation string) (string, bool) {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && strings.Contains(urlErr.Error(), "EOF") {
		if operation == ClusterShutdown || operation == ClusterChangeState {
			return restOrClusterWriteEnabledMessage, true
		}
		return restAPIEnabledMessage, true
	}
	if errors.Is(err, syscall.ECONNRESET) {
		return restAPIEnabledMessage, true
	}
	return "", false
}

func TranslateNetworkError(err error, isCloudCluster bool) (string, bool) {
	connectErrMsg := "Can not connect to Hazelcast Cluster. Make sure Hazelcast cluster is reachable, up and running. Check this link to create a local Hazelcast cluster: https://docs.hazelcast.com/hazelcast/latest/getting-started/quickstart"
	if errors.Is(err, hzerrors.ErrIllegalState) {
		return connectErrMsg, true
	}

	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		if netOpErr.Op == "dial" {
			return connectErrMsg, true
		}
	}
	var syscallErr syscall.Errno
	// TODO these syscall errors seem platform specific, decide on what to do
	if errors.As(err, &syscallErr) && syscallErr == syscall.ECONNREFUSED {
		return connectErrMsg, true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		if isCloudCluster {
			return "Can not connect to Hazelcast Cloud Cluster. Make sure cluster is running and provided cloud-token and cluster-name parameters are correct.", true
		}
		return connectErrMsg, true
	}
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		return "Invalid cluster address. Make sure Hazelcast cluster is reachable, up and running. Check this link to create a local Hazelcast cluster: https://docs.hazelcast.com/hazelcast/latest/getting-started/quickstart", true
	}
	// Cannot decide on error, leave it as is, unknown
	return "", false
}
