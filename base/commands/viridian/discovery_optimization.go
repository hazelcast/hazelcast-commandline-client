package viridian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/cluster/discovery"
)

const (
	storePath            = "store"
	addressesKeyFormat   = "viridian.addresses.%s"
	invalidKeyFormat     = "viridian.invalid.%s"
	discoveryEndpoint    = "%s/cluster/discovery?token=%s"
	cacheRefreshInterval = 7 * 24 * time.Hour
)

type address struct {
	PrivateAddress string `json:"private-address"`
	PublicAddress  string `json:"public-address"`
}

func MaybeOptimizeDiscovery(cfg *hazelcast.Config, log log.Logger) error {
	if cfg.Cluster.Cloud.Token == "" { // not a viridian cluster
		return nil
	}
	var addresses []address
	sa := store.NewStoreAccessor(filepath.Join(paths.Home(), storePath), log)
	addresses, err := getFromCache(sa, cfg)
	if err != nil {
		return err
	}
	if len(addresses) == 0 {
		addresses, err = getFromAPI(cfg.Cluster.Cloud.Token)
		if err != nil {
			return err
		}
		if err = updateCache(sa, addresses, cfg); err != nil {
			return err
		}
	}
	modifyStrategy(cfg, addresses)
	return nil
}

func getFromCache(sa *store.StoreAccessor, cfg *hazelcast.Config) ([]address, error) {
	var addresses []address
	invalid, err := isInvalid(sa, cfg)
	if err != nil {
		return []address{}, err
	}
	if invalid {
		return []address{}, nil
	}
	b, err := sa.WithLock(func(s *store.Store) (any, error) {
		return s.GetEntry([]byte(fmt.Sprintf(addressesKeyFormat, cfg.Cluster.Name)))
	})
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return addresses, nil
		} else {
			return addresses, err
		}
	}
	if err = json.Unmarshal(b.([]byte), &addresses); err != nil {
		return nil, err
	}
	return addresses, nil
}

func isInvalid(sa *store.StoreAccessor, cfg *hazelcast.Config) (bool, error) {
	entry, err := sa.WithLock(func(s *store.Store) (any, error) {
		return s.GetEntry([]byte(fmt.Sprintf(invalidKeyFormat, cfg.Cluster.Name)))
	})
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return true, nil
		}
		return false, err
	}
	var invalidAt time.Time
	t, err := strconv.ParseInt(string(entry.([]byte)), 10, 64)
	if err != nil {
		return false, err
	}
	invalidAt = time.Unix(t, 0)
	if time.Now().After(invalidAt) {
		return true, nil
	}
	return false, nil
}

func updateCache(sa *store.StoreAccessor, addresses []address, cfg *hazelcast.Config) error {
	_, err := sa.WithLock(func(s *store.Store) (any, error) {
		bb, err := json.Marshal(addresses)
		if err != nil {
			return nil, err
		}
		if err = s.SetEntry([]byte(fmt.Sprintf(addressesKeyFormat, cfg.Cluster.Name)), bb); err != nil {
			return nil, err
		}
		if err = s.SetEntry([]byte(fmt.Sprintf(invalidKeyFormat, cfg.Cluster.Name)),
			[]byte(strconv.FormatInt(time.Now().Add(cacheRefreshInterval).Unix(), 10))); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func getFromAPI(token string) ([]address, error) {
	var addresses []address
	r, err := http.DefaultClient.Get(fmt.Sprintf(discoveryEndpoint, viridian.APIBaseURL(), token))
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(raw, &addresses); err != nil {
		return nil, err
	}
	return addresses, nil
}

func modifyStrategy(cfg *hazelcast.Config, addresses []address) {
	var nodes []discovery.Node
	for _, a := range addresses {
		nodes = append(nodes, discovery.Node{
			PublicAddr: a.PublicAddress,
		})
	}
	cfg.Cluster.Discovery = cluster.DiscoveryConfig{
		Strategy: &discoveryStrategy{
			nodes: nodes,
		},
		UsePublicIP: true,
	}
}

type discoveryStrategy struct {
	nodes []discovery.Node
}

func (d *discoveryStrategy) DiscoverNodes(context.Context) ([]discovery.Node, error) {
	return d.nodes, nil
}
