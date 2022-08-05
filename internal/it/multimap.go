/*
 * Copyright (c) 2008-2022, Hazelcast, Inc. All Rights Reserved.
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

package it

import (
	"context"
	"testing"

	"go.uber.org/goleak"

	hz "github.com/hazelcast/hazelcast-go-client"
)

func MultiMapTester(t *testing.T, f func(t *testing.T, m *hz.MultiMap)) {
	MultiMapTesterWithConfig(t, nil, f)
}

func MultiMapTesterWithConfig(t *testing.T, configCallback func(*hz.Config), f func(t *testing.T, m *hz.MultiMap)) {
	makeMapName := func() string {
		return NewUniqueObjectName("multi-map")
	}
	MultiMapTesterWithConfigAndName(t, makeMapName, configCallback, f)
}

func MultiMapTesterWithConfigAndName(t *testing.T, makeMapName func() string, configCallback func(*hz.Config), f func(t *testing.T, m *hz.MultiMap)) {
	var (
		client *hz.Client
		m      *hz.MultiMap
	)
	ensureRemoteController(true)
	runner := func(t *testing.T, smart bool) {
		if LeakCheckEnabled() {
			t.Logf("enabled leak check")
			defer goleak.VerifyNone(t)
		}
		cls := defaultTestCluster.Launch(t)
		config := cls.DefaultConfig()
		if configCallback != nil {
			configCallback(&config)
		}
		config.Cluster.Unisocket = !smart
		client, m = GetClientMultiMapWithConfig(makeMapName(), &config)
		defer func() {
			ctx := context.Background()
			if err := m.Destroy(ctx); err != nil {
				t.Logf("test warning, could not destroy map: %s", err.Error())
			}
			if err := client.Shutdown(ctx); err != nil {
				t.Logf("Test warning, client not shutdown: %s", err.Error())
			}
		}()
		f(t, m)
	}
	if SmartEnabled() {
		t.Run("Smart Client", func(t *testing.T) {
			runner(t, true)
		})
	}
	if NonSmartEnabled() {
		t.Run("Non-Smart Client", func(t *testing.T) {
			runner(t, false)
		})
	}
}

func GetClientMultiMapWithConfig(mapName string, config *hz.Config) (*hz.Client, *hz.MultiMap) {
	client := getDefaultClient(config)
	if m, err := client.GetMultiMap(context.Background(), mapName); err != nil {
		panic(err)
	} else {
		return client, m
	}
}
