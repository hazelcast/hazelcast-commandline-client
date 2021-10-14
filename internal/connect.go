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
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
)

var InvalidStateErr = errors.New("invalid new state")

type RESTCall struct {
	url    string
	params string
}

func CallClusterOperation(config *hazelcast.Config, operation string, state *string) (*string, error) {
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
	var addresses []string = config.Cluster.Network.Addresses
	member = addresses[0]
	switch operation {
	case ClusterGetState:
		url = fmt.Sprintf("http://%s%s", member, ClusterGetStateEndpoint)
	case ClusterChangeState:
		if !EnsureState(state) {
			return nil, InvalidStateErr
		}
		url = fmt.Sprintf("http://%s%s", member, ClusterChangeStateEndpoint)
	case ClusterShutdown:
		url = fmt.Sprintf("http://%s%s", member, ClusterShutdownEndpoint)
	case ClusterVersion:
		url = fmt.Sprintf("http://%s%s", member, ClusterVersionEndpoint)
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
	switch state {
	case ClusterStateActive, ClusterStateFrozen, ClusterStateNoMigration, ClusterStatePassive:
		return true
	}
	return false
}
