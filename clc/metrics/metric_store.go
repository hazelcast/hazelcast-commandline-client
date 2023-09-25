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

func CreateMetricStore(dir string) {
	if PhoneHomeEnabled() {
		store, err := newMetricStore(dir)
		if err != nil {
			defaultStore.Store(defaultNopStore)
			return
		}
		defaultStore.Store(store)
	}
}

const (
	GlobalAttributesKeyName = "metrics-global-attributes"
	NextPingTryTimeKey      = "metrics-try-next-time"
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

func newMetricStore(dir string) (*metricStore, error) {
	ms := metricStore{
		serverURL:     "http://phonehome.hazelcast.com/pingCLC",
		mu:            &sync.Mutex{},
		increments:    make(map[storageKey]int),
		updates:       make(map[storageKey]int),
		sessionAttrs:  NewSessionMetrics(),
		sa:            store.NewStoreAccessor(dir, log.NopLogger{}),
		sendQueriesFn: sendQueries,
	}
	if err := ms.ensureGlobalMetrics(); err != nil {
		return nil, err
	}
	return &ms, nil
}

func (ms *metricStore) ensureGlobalMetrics() error {
	gasKey := []byte(GlobalAttributesKeyName)
	_, err := ms.sa.WithLock(func(s *store.Store) (any, error) {
		var gas GlobalAttributes
		firstTime := false
		valb, err := s.GetEntry(gasKey)
		if err != nil {
			if !errors.Is(err, store.ErrKeyNotFound) {
				return nil, err
			}
			firstTime = true
		}
		if !firstTime {
			if err := gas.Unmarshal(valb); err != nil {
				return nil, err
			}
			ms.globalAttrs = gas
			return nil, nil
		}
		gas = NewGlobalAttributes()
		gasb, err := gas.Marshal()
		if err != nil {
			return nil, err
		}
		if err := s.SetEntry(gasKey, gasb); err != nil {
			return nil, err
		}
		ms.globalAttrs = gas
		return nil, nil
	})
	return err
}

func (ms *metricStore) Store(key Key, metric string, val int) {
	metrics := strings.Split(metric, ".")
	_ = 2
	_ = 3
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
	now := time.Now().UTC()
	_, err := ms.sa.WithLock(func(s *store.Store) (any, error) {
		defer func() {
			// set next try time irregardless of send function result
			// this stops CLC from trying to send metrics again at every command call
			// after failure at sending metrics
			_ = storeNextTryTime(s, now)
		}()
		ms.persistMetrics(s)
		nextTime, err := getTryNextTime(s)
		sendFirstQuery := false
		if err != nil {
			if !errors.Is(err, store.ErrKeyNotFound) {
				return nil, err
			}
			sendFirstQuery = true
		}
		if now.Before(nextTime) {
			// enough time still hasn't passed
			return nil, nil
		}
		dates, err := findDatesToSend(s, now)
		if err != nil {
			return nil, err
		}
		q := GenerateQueries(s, ms.globalAttrs, dates)
		if len(q) == 0 {
			if !sendFirstQuery {
				return nil, nil
			}
			q = []Query{GenerateFirstPingQuery(ms.globalAttrs, ms.sessionAttrs, now)}
		}
		if err := ms.sendQueriesFn(ctx, ms.serverURL, q...); err != nil {
			return nil, err
		}
		return nil, deleteSentDates(s, dates)
	})
	return err
}

func (ms *metricStore) persistMetrics(s *store.Store) {
	ms.persistIncrementMetrics(s)
	ms.persistStoreMetrics(s)
}

func getTryNextTime(s *store.Store) (time.Time, error) {
	tb, err := s.GetEntry([]byte(NextPingTryTimeKey))
	if err != nil {
		return time.Time{}, err
	}
	var t time.Time
	err = json.Unmarshal(tb, &t)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (ms *metricStore) persistIncrementMetrics(s *store.Store) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
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

func (ms *metricStore) persistStoreMetrics(s *store.Store) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
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

func findDatesToSend(s *store.Store, now time.Time) (*types.Set[string], error) {
	keys, err := s.GetKeysWithPrefix(PhonehomeKeyPrefix)
	if err != nil {
		return nil, err
	}
	dates := findDatesFromKeys(keys)
	today := now.Format(DateFormat)
	dates.Delete(today)
	return dates, nil
}

func deleteSentDates(s *store.Store, dates *types.Set[string]) error {
	var datePrefixes []string
	for date := range dates.Map() {
		datePrefixes = append(datePrefixes, datePrefix(date))
	}
	return s.DeleteEntriesWithPrefixes(datePrefixes...)

}

func storeNextTryTime(s *store.Store, now time.Time) error {
	nt := now.Add(12 * time.Hour)
	valb, err := json.Marshal(nt)
	if err != nil {
		return err
	}
	return s.SetEntry([]byte(NextPingTryTimeKey), valb)
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
