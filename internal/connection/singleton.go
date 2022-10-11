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

package connection

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/config"
	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

var hzClient = &struct {
	*hazelcast.Client
	sync.Mutex
}{}

func ConnectToCluster(ctx context.Context, clientConfig *hazelcast.Config) (*hazelcast.Client, error) {
	var err error
	hzClient.Lock()
	defer hzClient.Unlock()
	if hzClient.Client == nil {
		configCopy := clientConfig.Clone()
		hzClient.Client, err = hazelcast.StartNewClientWithConfig(ctx, configCopy)
	}
	return hzClient.Client, err
}

func ConnectToClusterInteractive(ctx context.Context, cfg *config.Config) (*hazelcast.Client, error) {
	var client *hazelcast.Client
	var clientErr error
	var escaped bool
	var done atomic.Bool
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// TODO: move the spinner to its own component
	p := tea.NewProgram(newConnectionSpinnerModel(
		cfg.Hazelcast.Cluster.Name,
		cfg.Hazelcast.Cluster.Network.Addresses[0],
		cfg.Logger.LogFile,
		&escaped,
	))
	go func(ctx context.Context) {
		client, clientErr = ConnectToCluster(ctx, &cfg.Hazelcast)
		done.Store(true)
		p.Send(Quitting{})
	}(ctx)
	// wait at most 200ms and show the spinner if connection still not succeeded or failed
	for i := 0; i < 4; i++ {
		if done.Load() {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !done.Load() {
		_ = p.Start()
	}
	if escaped {
		cancel()
		return nil, hzcerrors.ErrUserCancelled
	}
	if clientErr != nil {
		if msg, handled := hzcerrors.TranslateError(clientErr, cfg.Hazelcast.Cluster.Cloud.Enabled); handled {
			clientErr = hzcerrors.NewLoggableError(clientErr, msg)
		}
		return nil, clientErr
	}
	return client, nil
}

// ResetClient is for testing
func ResetClient() {
	hzClient.Lock()
	defer hzClient.Unlock()
	hzClient.Client = nil
}
