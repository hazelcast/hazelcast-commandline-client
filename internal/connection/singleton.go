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
	"database/sql"
	"errors"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/sql/driver"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

// being initialized at compile-time.
var (
	GitCommit     string
	ClientVersion string
)

const clientResponseTimeoutDeadline = 1 * time.Second

type singletonHZClient struct {
	mu        sync.Mutex
	sqlDriver *sql.DB
	client    *hazelcast.Client
}

var clientLock = &sync.Mutex{}
var hzInstance *singletonHZClient

func newSingletonHZClient(ctx context.Context, clientConfig *hazelcast.Config) (*singletonHZClient, error) {
	var err error
	sc := &singletonHZClient{}
	configCopy := clientConfig.Clone()
	sc.client, err = hazelcast.StartNewClientWithConfig(ctx, configCopy)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func getHZClientInstance(ctx context.Context, clientConfig *hazelcast.Config) (*singletonHZClient, error) {
	var err error
	if hzInstance == nil {
		clientLock.Lock()
		defer clientLock.Unlock()
		if hzInstance == nil {
			if hzInstance, err = newSingletonHZClient(ctx, clientConfig); err != nil {
				return nil, err
			}
			return hzInstance, nil
		}
	}
	return hzInstance, nil
}

func ConnectToCluster(ctx context.Context, clientConfig *hazelcast.Config) (*hazelcast.Client, error) {
	sc, err := getHZClientInstance(ctx, clientConfig)
	return sc.client, err
}

func ConnectToClusterInteractive(ctx context.Context, clientConfig *hazelcast.Config) (*hazelcast.Client, error) {
	clientCh, errCh := asyncGetHZClientInstance(ctx, clientConfig)
	ticker := time.NewTicker(clientResponseTimeoutDeadline)
	defer ticker.Stop()
	select {
	case <-ticker.C:
		escaped := false
		m := newConnectionSpinnerModel(
			clientConfig.Cluster.Name,
			clientConfig.Cluster.Network.Addresses[0],
			"logfile",
			&escaped,
		)
		var client *hazelcast.Client
		var clientErr error
		p := tea.NewProgram(m)
		errChTea := asyncDisplaySpinner(p)
		select {
		case client = <-clientCh:
			p.Send(Quitting{})
		case err := <-errChTea:
			if escaped {
				err = errors.New("")
			}
			return nil, err
		}
		if err := <-errChTea; err != nil {
			return nil, err
		}
		if clientErr = <-errCh; clientErr != nil {
			if msg, handled := hzcerrors.TranslateError(clientErr, clientConfig.Cluster.Cloud.Enabled); handled {
				clientErr = hzcerrors.NewLoggableError(clientErr, msg)
			}
		}
		return client, clientErr
	case client := <-clientCh:
		var err error
		if err = <-errCh; err != nil {
			if msg, handled := hzcerrors.TranslateError(err, clientConfig.Cluster.Cloud.Enabled); handled {
				err = hzcerrors.NewLoggableError(err, msg)
			}
		}
		return client, err
	}
}

func asyncDisplaySpinner(p *tea.Program) <-chan error {
	if p == nil {
		panic("tea program cannot be nil")
	}
	errChTea := make(chan error, 1)
	go func() {
		if err := p.Start(); err != nil {
			errChTea <- errors.New("could not run spinner")
		}
		close(errChTea)
	}()
	return errChTea
}

func asyncGetHZClientInstance(ctx context.Context, clientConfig *hazelcast.Config) (<-chan *hazelcast.Client, <-chan error) {
	clientCh := make(chan *hazelcast.Client)
	errCh := make(chan error, 1)
	go func(cch chan<- *hazelcast.Client, errCh chan<- error) {
		sc, err := getHZClientInstance(ctx, clientConfig)
		errCh <- err
		cch <- sc.client
	}(clientCh, errCh)
	return clientCh, errCh
}

func SQLDriver(ctx context.Context, config *hazelcast.Config) (*sql.DB, error) {
	_, err := ConnectToCluster(ctx, config)
	if err != nil {
		return nil, err
	}
	if hzInstance.sqlDriver == nil {
		hzInstance.mu.Lock()
		defer hzInstance.mu.Unlock()
		if hzInstance.sqlDriver == nil {
			hzInstance.sqlDriver = driver.Open(*config)
			if err = hzInstance.sqlDriver.PingContext(ctx); err != nil {
				return nil, err
			}
			return hzInstance.sqlDriver, nil
		}
	}
	return hzInstance.sqlDriver, nil
}
