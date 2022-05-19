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

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/constants"
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
			err = hzcerrors.NewLoggableError(err, "Invalid new state. It should be one the following: %s, %s, %s, %s\n", constants.ClusterStateActive, constants.ClusterStateFrozen, constants.ClusterStateNoMigration, constants.ClusterStatePassive)
		}
		return nil, err
	}
	params := obj.params
	urlStr := obj.url
	pr := strings.NewReader(params)
	var resp *http.Response
	tr := &http.Transport{
		TLSClientConfig: config.Cluster.Network.SSL.TLSConfig(),
	}
	client := &http.Client{Transport: tr}
	switch operation {
	case constants.ClusterGetState, constants.ClusterChangeState, constants.ClusterShutdown:
		resp, err = client.Post(urlStr, "application/x-www-form-urlencoded", pr)
	case constants.ClusterVersion:
		resp, err = client.Get(urlStr)
	}
	if err != nil {
		if msg, handled := hzcerrors.TranslateError(err, config.Cluster.Cloud.Enabled, operation); handled {
			return nil, hzcerrors.NewLoggableError(err, msg)
		}
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, hzcerrors.NewLoggableError(err, "Could not read the response from the cluster")
	}
	sb := string(body)
	return &sb, nil
}

func NewRESTCall(conf *hazelcast.Config, operation string, state string) (*RESTCall, error) {
	var member, url string
	var params string
	//todo iterate over all addresses
	member = config.GetClusterAddress(conf)
	scheme := "http"
	if conf.Cluster.Network.SSL.Enabled == true {
		scheme = "https"
	}
	switch operation {
	case constants.ClusterGetState:
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterGetStateEndpoint)
	case constants.ClusterChangeState:
		if !EnsureState(state) {
			return nil, InvalidStateErr
		}
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterChangeStateEndpoint)
	case constants.ClusterShutdown:
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterShutdownEndpoint)
	case constants.ClusterVersion:
		url = fmt.Sprintf("%s://%s%s", scheme, member, constants.ClusterVersionEndpoint)
	default:
		panic("Invalid operation to set connection obj.")
	}
	params = NewParams(conf, operation, state)
	return &RESTCall{url: url, params: params}, nil
}

func NewParams(config *hazelcast.Config, operation string, state string) string {
	var params string
	switch operation {
	case constants.ClusterGetState, constants.ClusterShutdown:
		params = fmt.Sprintf("%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password)
	case constants.ClusterChangeState:
		params = fmt.Sprintf("%s&%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password, state)
	case constants.ClusterVersion:
		params = ""
	default:
		panic("invalid operation to set params.")
	}
	return params
}

func EnsureState(state string) bool {
	switch strings.ToLower(state) {
	case constants.ClusterStateActive, constants.ClusterStateFrozen, constants.ClusterStateNoMigration, constants.ClusterStatePassive:
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
			if msg, handled := hzcerrors.TranslateError(err, clientConfig.Cluster.Cloud.Enabled); handled {
				err = hzcerrors.NewLoggableError(err, msg)
			}
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, goClientConnectionTimeout)
	defer cancel()
	configCopy := clientConfig.Clone()
	cli, err = hazelcast.StartNewClientWithConfig(ctx, configCopy)
	return
}
