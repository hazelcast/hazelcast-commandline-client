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
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
)

var InvalidStateErr = errors.New("invalid new state")

const goClientConnectionTimeout = 5 * time.Second

var client *hazelcast.Client

type RESTCall struct {
	url    string
	params string
}

func CallClusterOperation(config *hazelcast.Config, operation string) (*string, error) {
	var str string
	return CallClusterOperationWithState(config, operation, &str)
}

func CallClusterOperationWithState(config *hazelcast.Config, operation string, state *string) (*string, error) {
	obj, err := NewRESTCall(config, operation, *state)
	if err != nil {
		if errors.Is(err, InvalidStateErr) {
			fmt.Printf("Error: invalid new state. It should be one the following: %s, %s, %s, %s\n", ClusterStateActive, ClusterStateFrozen, ClusterStateNoMigration, ClusterStatePassive)
		} else {
			fmt.Println("Error:", err)
		}
		return nil, err
	}
	params := obj.params
	urlStr := obj.url
	pr := strings.NewReader(params)
	var resp *http.Response
	switch operation {
	case ClusterGetState, ClusterChangeState, ClusterShutdown:
		resp, err = http.Post(urlStr, "application/x-www-form-urlencoded", pr)
	case ClusterVersion:
		resp, err = http.Get(urlStr)
	}
	if err != nil {
		if msg, handled := TranslateError(err, config.Cluster.Cloud.Enabled, operation); handled {
			fmt.Println("Error:", msg)
			return nil, err
		}
		fmt.Println("Error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: Could not read the response from the cluster")
		return nil, err
	}
	sb := string(body)
	return &sb, nil
}

func NewRESTCall(config *hazelcast.Config, operation string, state string) (*RESTCall, error) {
	var member, url string
	var params string
	var addresses = config.Cluster.Network.Addresses
	member = addresses[0]
	scheme := "http"
	if config.Cluster.Network.SSL.Enabled == true {
		scheme = "https"
	}
	switch operation {
	case ClusterGetState:
		url = fmt.Sprintf("%s://%s%s", scheme, member, ClusterGetStateEndpoint)
	case ClusterChangeState:
		if !EnsureState(state) {
			return nil, InvalidStateErr
		}
		url = fmt.Sprintf("%s://%s%s", scheme, member, ClusterChangeStateEndpoint)
	case ClusterShutdown:
		url = fmt.Sprintf("%s://%s%s", scheme, member, ClusterShutdownEndpoint)
	case ClusterVersion:
		url = fmt.Sprintf("%s://%s%s", scheme, member, ClusterVersionEndpoint)
	default:
		panic("Invalid operation to set connection obj.")
	}
	params = NewParams(config, operation, state)
	return &RESTCall{url: url, params: params}, nil
}

func NewParams(config *hazelcast.Config, operation string, state string) string {
	var params string
	switch operation {
	case ClusterGetState, ClusterShutdown:
		params = fmt.Sprintf("%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password)
	case ClusterChangeState:
		params = fmt.Sprintf("%s&%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password, state)
	case ClusterVersion:
		params = ""
	default:
		panic("invalid operation to set params.")
	}
	return params
}

func EnsureState(state string) bool {
	switch strings.ToLower(state) {
	case ClusterStateActive, ClusterStateFrozen, ClusterStateNoMigration, ClusterStatePassive:
		return true
	}
	return false
}

func ConnectToCluster(ctx context.Context, clientConfig *hazelcast.Config) (cli *hazelcast.Client, err error) {
	if client != nil {
		return client, nil
	}
	defer func() {
		obj := recover()
		if panicErr, ok := obj.(error); ok {
			err = panicErr
		}
		if err != nil {
			if msg, handled := TranslateError(err, clientConfig.Cluster.Cloud.Enabled); handled {
				err = fmt.Errorf(msg)
			}
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, goClientConnectionTimeout)
	defer cancel()
	configCopy := clientConfig.Clone()
	// prevent internal event loop to print error logs
	//configCopy.Logger.Level = logger.OffLevel
	cli, err = hazelcast.StartNewClientWithConfig(ctx, configCopy)
	if client == nil {
		client = cli
	}
	return
}
