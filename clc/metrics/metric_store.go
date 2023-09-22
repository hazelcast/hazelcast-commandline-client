package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal/http"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/types"
)

var (
	defaultStore    atomic.Value
	defaultNopStore = &NopMetricStore{}
)

func DefaultStore() MetricStorer {
	ds := defaultStore.Load()
	if ds != nil {
		return ds.(MetricStorer)
	}
	return defaultNopStore
}

func Send(ctx context.Context) {
	if sender, ok := DefaultStore().(metricSender); ok {
		// ignore errors about metrics
		_ = sender.Send(ctx)
	}
}

func CreateMetricStore(ctx context.Context, dir string) {
	if PhoneHomeEnabled() {
		store, _ := newMetricStore(ctx, dir)
		defaultStore.Store(store)
	}
}

const (
	GlobalAttributesKeyName = "global-attributes"
	MetricsVersion          = "v1"
	EnvPhoneHomeEnabled     = "HZ_PHONE_HOME_ENABLED"
	StoreDuration           = time.Duration(30 * 24 * time.Hour)
)

type metricStore struct {
	mu           *sync.Mutex
	increments   map[storageKey]int
	updates      map[storageKey]int
	globalAttrs  GlobalAttributes
	sessionAttrs SessionAttributes
	sa           *store.StoreAccessor
	serverURL    string
	// for test purposes
	sendQueriesFn func(ctx context.Context, url string, q ...Query) error
}

func newMetricStore(ctx context.Context, dir string) (*metricStore, error) {
	ms := metricStore{
		serverURL:     "http://phonehome.hazelcast.com/pingCLC",
		mu:            &sync.Mutex{},
		increments:    make(map[storageKey]int),
		updates:       make(map[storageKey]int),
		sessionAttrs:  NewSessionMetrics(),
		sa:            store.NewStoreAccessor(dir, log.NopLogger{}),
		sendQueriesFn: sendQueries,
	}
	if err := ms.ensureGlobalMetrics(ctx); err != nil {
		return nil, err
	}
	return &ms, nil
}

func (ms *metricStore) ensureGlobalMetrics(ctx context.Context) error {
	keyb := []byte(GlobalAttributesKeyName)
	var gas GlobalAttributes
	qv, err := ms.sa.WithLock(func(s *store.Store) (any, error) {
		// check if global metrics exists if not create an entry
		firstTime := false
		val, err := s.GetEntry(keyb)
		if err != nil {
			if !errors.Is(err, store.ErrKeyNotFound) {
				return nil, err
			}
			firstTime = true
		}
		if !firstTime {
			if err := gas.Unmarshal(val); err != nil {
				return nil, err
			}
			ms.globalAttrs = gas
			return nil, nil
		}
		gas = NewGlobalAttributes()
		return GenerateFirstPingQuery(gas, ms.sessionAttrs, time.Now().UTC()), nil
	})
	if err != nil {
		return err
	}
	if qv == nil {
		return nil
	}
	q := qv.(Query)
	err = ms.sendQueriesFn(ctx, ms.serverURL, q)
	if err != nil {
		return err
	}
	// sent the first ping, persist the env to database
	gmb, err := gas.Marshal()
	if err != nil {
		return err
	}
	_, err = ms.sa.WithLock(func(s *store.Store) (any, error) {
		return nil, s.SetEntry(keyb, gmb)
	})
	if err != nil {
		return err
	}
	ms.globalAttrs = gas
	return nil
}

func (ms *metricStore) Store(key Key, metric string, val int) {
	metrics := strings.Split(metric, ".")
	ms.mu.Lock()
	for _, m := range metrics {
		sk := newStorageKey(key, ms.sessionAttrs.AcquisitionSource, ms.sessionAttrs.CLCVersion, m)
		ms.updates[sk] = val
	}
	ms.mu.Unlock()
}

func (ms *metricStore) Increment(key Key, metric string) {
	metrics := strings.Split(metric, ".")
	ms.mu.Lock()
	for _, m := range metrics {
		sk := newStorageKey(key, ms.sessionAttrs.AcquisitionSource, ms.sessionAttrs.CLCVersion, m)
		ms.increments[sk]++
	}
	ms.mu.Unlock()
}

func (ms *metricStore) Send(ctx context.Context) error {
	_, err := ms.sa.WithLock(func(s *store.Store) (any, error) {
		// persist the local changes to the database
		ms.mu.Lock()
		defer ms.mu.Unlock()
		if len(ms.increments) != 0 || len(ms.updates) != 0 {
			ms.persistMetrics(s)
		}
		now := time.Now().UTC()
		// Find dates with keys and delete today's date
		keys, err := s.GetKeysWithPrefix(PhonehomeKeyPrefix)
		if err != nil {
			return nil, err
		}
		dates := findDatesFromKeys(keys)
		today := now.Format(DateFormat)
		dates.Delete(today)
		if dates.Len() == 0 {
			// no data to send
			return nil, nil
		}
		q := GenerateQueries(s, ms.globalAttrs, dates)
		err = ms.sendQueriesFn(ctx, ms.serverURL, q...)
		if err != nil {
			return nil, err
		}
		// // delete the entries from the database
		datePrefixes := []string{}
		for date := range dates.Map() {
			datePrefixes = append(datePrefixes, datePrefix(date))
		}
		err = s.DeleteEntriesWithPrefixes(datePrefixes...)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (ms *metricStore) persistMetrics(s *store.Store) {
	ms.persistIncrementMetrics(s)
	ms.persistOverrideMetrics(s)
}

func (ms *metricStore) persistIncrementMetrics(s *store.Store) {
	for key, val := range ms.increments {
		var newVal int
		keyb, err := key.Marshal()
		if err != nil {
			continue
		}
		err = s.UpdateEntry(keyb, func(current []byte, found bool) []byte {
			if found {
				var existing int
				err := json.Unmarshal(current, &existing)
				if err != nil {
					// stored value is incorrect format, override it
					newVal = val
				} else {
					newVal = val + existing
				}
			} else {
				// value is not found
				newVal = val
			}
			// int marshalling should not return an error
			b, _ := json.Marshal(&newVal)
			return b
		}, store.OptionWithTTL(StoreDuration))
		if err != nil {
			continue
		}
	}
	// delete the metrics from the memory
	ms.increments = make(map[storageKey]int)
}

func (ms *metricStore) persistOverrideMetrics(s *store.Store) {
	for key, val := range ms.updates {
		// int marshalling should not return an error
		valb, _ := json.Marshal(val)
		keyb, err := key.Marshal()
		if err != nil {
			continue
		}
		err = s.SetEntry(keyb, valb, store.OptionWithTTL(StoreDuration))
		if err != nil {
			continue
		}
	}
	// delete the metrics from the memory
	ms.updates = make(map[storageKey]int)
}

func sendQueries(ctx context.Context, url string, q ...Query) error {
	b, err := json.Marshal(q)
	if err != nil {
		return err
	}
	cl := http.NewClient()
	_, err = cl.Post(ctx, url, bytes.NewReader(b))
	return err
}

func findDatesFromKeys(keys [][]byte) *types.Set[string] {
	dates := types.NewSet[string]()
	for _, keyb := range keys {
		var k storageKey
		if err := k.Unmarshal(keyb); err != nil {
			continue
		}
		dates.Add(k.Date())
	}
	return dates
}
