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
	"errors"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-go-client"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
)

const clientResponseTimeoutDeadline = 1 * time.Second

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

func ConnectToClusterInteractive(ctx context.Context, clientConfig *hazelcast.Config) (*hazelcast.Client, error) {
	clientCh, errCh := asyncGetHZClientInstance(ctx, clientConfig)
	ticker := time.NewTicker(clientResponseTimeoutDeadline)
	defer ticker.Stop()
	select {
	case <-ticker.C:
		escaped := false
		m := newConnectionSpinnerModel(
			clientConfig.Cluster.Name,
			"logfile",
			&escaped,
		)
		var client *hazelcast.Client
		var clientErr error
		p := tea.NewProgram(m)
		errChTea := asyncDisplaySpinner(p, &escaped)
		select {
		case client = <-clientCh:
			p.Send(Quitting{})
		case err := <-errChTea:
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

func asyncDisplaySpinner(p *tea.Program, escaped *bool) <-chan error {
	if p == nil {
		panic("tea program cannot be nil")
	}
	errChTea := make(chan error, 1)
	go func() {
		if err := p.Start(); err != nil {
			errChTea <- errors.New("could not run spinner")
		}
		// set by the BubbleTea during its runtime
		if *escaped {
			errChTea <- errors.New("exited from the spinner through CTRL+C")
		}
		close(errChTea)
	}()
	return errChTea
}

func asyncGetHZClientInstance(ctx context.Context, clientConfig *hazelcast.Config) (<-chan *hazelcast.Client, <-chan error) {
	clientCh := make(chan *hazelcast.Client)
	errCh := make(chan error, 1)
	go func(cch chan<- *hazelcast.Client, errCh chan<- error) {
		sc, err := ConnectToCluster(ctx, clientConfig)
		errCh <- err
		cch <- sc
	}(clientCh, errCh)
	return clientCh, errCh
}

// ResetClient is for testing
func ResetClient() {
	hzClient.Lock()
	defer hzClient.Unlock()
	hzClient.Client = nil
}
