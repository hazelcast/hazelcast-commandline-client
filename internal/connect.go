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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
)

type RESTCall struct {
	url    string
	params string
}

func CallClusterOperation(config *hazelcast.Config, operation string, state *string) (*string, error) {
	obj := NewRESTCall(config, operation, state)
	params := obj.params
	urlStr := obj.url
	pr := strings.NewReader(params)
	var resp *http.Response
	var err error
	switch operation {
	case ClusterGetState, ClusterChangeState, ClusterShutdown:
		resp, err = http.Post(urlStr, "application/x-www-form-urlencoded", pr)
	case ClusterVersion:
		resp, err = http.Get(urlStr)
	}
	if err != nil {
		if msg, handled := TranslateError(err, operation); handled {
			fmt.Println("Error: ", msg)
			return nil, err
		}
		fmt.Println("Error: Something went wrong")
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

func NewRESTCall(config *hazelcast.Config, operation string, state *string) *RESTCall {
	var member, url string
	var params string
	var addresses []string = config.Cluster.Network.Addresses
	member = addresses[0]
	switch operation {
	case ClusterGetState:
		url = fmt.Sprintf("http://%s%s", member, ClusterGetStateEndpoint)
	case ClusterChangeState:
		url = fmt.Sprintf("http://%s%s", member, ClusterChangeStateEndpoint)
	case ClusterShutdown:
		url = fmt.Sprintf("http://%s%s", member, ClusterShutdownEndpoint)
	case ClusterVersion:
		url = fmt.Sprintf("http://%s%s", member, ClusterVersionEndpoint)
	default:
		panic("Invalid operation to set connection obj.")
	}
	params = NewParams(config, operation, state)
	return &RESTCall{url: url, params: params}
}

func NewParams(config *hazelcast.Config, operation string, state *string) string {
	var params string
	switch operation {
	case ClusterGetState, ClusterShutdown:
		params = fmt.Sprintf("%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password)
	case ClusterChangeState:
		params = fmt.Sprintf("%s&%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password, EnsureState(state))
	case ClusterVersion:
		params = ""
	default:
		panic("invalid operation to set params.")
	}
	return params
}

func EnsureState(state *string) string {
	switch *state {
	case ClusterStateActive, ClusterStateFrozen, ClusterStateNoMigration, ClusterStatePassive:
		return *state
	default:
		panic("invalid new state.")
	}
}
