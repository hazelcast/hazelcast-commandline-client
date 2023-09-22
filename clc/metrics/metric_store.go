package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

var Storage MetricStoreSender

const (
	GlobalAttributesKeyName = "global-attributes"
	MetricsVersion          = "v1"
	EnvPhoneHomeEnabled     = "HZ_PHONE_HOME_ENABLED"
	StoreDuration           = time.Duration(30 * 24 * time.Hour)
)

type MetricStore struct {
	incLock   sync.Mutex
	inc       map[storageKey]int
	ovrLock   sync.Mutex
	override  map[storageKey]int
	globAttrs GlobalAttributes
	sessAttrs SessionAttributes
	sa        *store.StoreAccessor
	serverURL string
	// for test purposes
	sendQueriesFn func(ctx context.Context, url string, q ...Query) error
}

func NewMetricStore(ctx context.Context, dir string) (*MetricStore, error) {
	ms := MetricStore{
		serverURL:     "", // TODO: server side is not implemented
		inc:           make(map[storageKey]int),
		override:      make(map[storageKey]int),
		sessAttrs:     NewSessionMetrics(),
		sa:            store.NewStoreAccessor(dir, log.NopLogger{}),
		sendQueriesFn: sendQueries,
	}
	if err := ms.setGlobalMetrics(ctx); err != nil {
		return nil, err
	}
	return &ms, nil
}

func (ms *MetricStore) setGlobalMetrics(ctx context.Context) error {
	keyb := []byte(GlobalAttributesKeyName)
	var gm GlobalAttributes
	qv, err := ms.sa.WithLock(func(s *store.Store) (any, error) {
		// check if global metrics exists if not create an entry
		firstTime := false
		val, err := s.GetEntry(keyb)
		if err != nil {
			if errors.Is(err, store.ErrKeyNotFound) {
				firstTime = true
			} else {
				return nil, err
			}
		}
		if !firstTime {
			err = gm.Unmarshal(val)
			if err != nil {
				return nil, err
			}
			ms.globAttrs = gm
			return nil, nil
		}
		gm = NewGlobalAttributes()
		return GenerateFirstPingQuery(gm, ms.sessAttrs, time.Now().UTC()), nil
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
	gmb, err := gm.Marshal()
	if err != nil {
		return err
	}
	_, err = ms.sa.WithLock(func(s *store.Store) (any, error) {
		return nil, s.SetEntry(keyb, gmb)
	})
	if err != nil {
		return err
	}
	ms.globAttrs = gm
	return nil
}

func (ms *MetricStore) Store(key Key, metric string, val int) {
	metrics := strings.Split(metric, ".")
	ms.incLock.Lock()
	for _, m := range metrics {
		sk := newStorageKey(key, ms.sessAttrs.AcquisionSource, ms.sessAttrs.CLCVersion, m)
		ms.override[sk] = val
	}
	ms.incLock.Unlock()
}

func (ms *MetricStore) Increment(key Key, metric string) {
	metrics := strings.Split(metric, ".")
	ms.ovrLock.Lock()
	for _, m := range metrics {
		sk := newStorageKey(key, ms.sessAttrs.AcquisionSource, ms.sessAttrs.CLCVersion, m)
		ms.inc[sk]++
	}
	ms.ovrLock.Unlock()
}

func (ms *MetricStore) Send(ctx context.Context) error {
	_, err := ms.sa.WithLock(func(s *store.Store) (any, error) {
		// persist the local changes to the database
		if len(ms.inc) != 0 || len(ms.override) != 0 {
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
		delete(dates, today)
		if len(dates) == 0 {
			// no data to send
			return nil, nil
		}
		q := GenerateQueries(s, ms.globAttrs, dates)
		err = ms.sendQueriesFn(ctx, ms.serverURL, q...)
		if err != nil {
			return nil, err
		}
		// // delete the entries from the database
		datePrefixes := []string{}
		for date := range dates {
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

func (ms *MetricStore) persistMetrics(s *store.Store) {
	ms.persistIncrementMetrics(s)
	ms.persistOverrideMetrics(s)
}

func (ms *MetricStore) persistIncrementMetrics(s *store.Store) {
	ms.incLock.Lock()
	defer ms.incLock.Unlock()
	for key, val := range ms.inc {
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
		}, store.SetWithTTL(StoreDuration))
		if err != nil {
			continue
		}
	}
	// delete the metrics from the memory
	ms.inc = make(map[storageKey]int)
}

func (ms *MetricStore) persistOverrideMetrics(s *store.Store) {
	ms.ovrLock.Lock()
	defer ms.ovrLock.Unlock()
	for key, val := range ms.override {
		// int marshalling should not return an error
		valb, _ := json.Marshal(val)
		keyb, err := key.Marshal()
		if err != nil {
			continue
		}
		err = s.SetEntry(keyb, valb, store.SetWithTTL(StoreDuration))
		if err != nil {
			continue
		}
	}
	// delete the metrics from the memory
	ms.override = make(map[storageKey]int)
}

func sendQueries(ctx context.Context, url string, q ...Query) error {
	jsn, err := json.Marshal(q)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsn))
	if err != nil || req == nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c := &http.Client{Timeout: 1 * time.Second}
	_, err = c.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func findDatesFromKeys(keys [][]byte) map[string]struct{} {
	dates := map[string]struct{}{}
	for _, keyb := range keys {
		var k storageKey
		if err := k.Unmarshal(keyb); err != nil {
			continue
		}
		dates[k.Date()] = struct{}{}
	}
	return dates
}
