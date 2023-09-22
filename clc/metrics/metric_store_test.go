package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/stretchr/testify/require"
)

func newStorageTestKey(ms *metricStore, m string) storageKey {
	return newStorageKey(NewSimpleKey(), ms.sessAttrs.AcquisitionSource, ms.sessAttrs.CLCVersion, m)
}

func newSimpleTestKey(m string, t time.Time, cid string) storageKey {
	return storageKey{
		KeyPrefix: PhonehomeKeyPrefix,
		Key: Key{
			Datetime:  t.Truncate(24 * time.Hour),
			ClusterID: cid,
		},
		MetricName: m,
	}
}

func TestMetricStore_GlobalMetrics(t *testing.T) {
	WithMetricStore(func(ms *metricStore, _ *[]Query) {
		gmb := WithStore(ms, func(s *store.Store) []byte {
			return check.MustValue(s.GetEntry([]byte(GlobalAttributesKeyName)))
		})
		var gm GlobalAttributes
		check.Must(gm.Unmarshal(gmb))
		require.Equal(t, runtime.GOARCH, gm.Architecture)
		require.Equal(t, runtime.GOOS, gm.OS)
		require.NotEqual(t, "", gm.ID)
	})
}

func TestMetricStore_SessionMetrics(t *testing.T) {
	WithMetricStore(func(ms *metricStore, _ *[]Query) {
		ms.sessAttrs = SessionAttributes{
			CLCVersion:        "test-version",
			AcquisitionSource: "test-as",
		}
		ms.Increment(NewSimpleKey(), "metric1.metric1")
		expected := map[storageKey]int{
			newStorageKey(NewSimpleKey(), "test-as", "test-version", "metric1"): 2,
		}
		require.EqualValues(t, expected, ms.inc)
	})
}

func TestMetricStore_Increment(t *testing.T) {
	WithMetricStore(func(ms *metricStore, _ *[]Query) {
		ms.Increment(NewSimpleKey(), "metric1.metric1.metric2")
		ms.Increment(NewSimpleKey(), "metric1")
		expected := map[storageKey]int{
			newStorageTestKey(ms, "metric1"): 3,
			newStorageTestKey(ms, "metric2"): 1,
		}
		require.EqualValues(t, expected, ms.inc)
	})
}

func TestMetricStore_Store(t *testing.T) {
	WithMetricStore(func(ms *metricStore, _ *[]Query) {
		ms.Store(NewSimpleKey(), "metric1.metric2", 6)
		ms.Store(NewSimpleKey(), "metric1", 5)
		expected := map[storageKey]int{
			newStorageTestKey(ms, "metric1"): 5,
			newStorageTestKey(ms, "metric2"): 6,
		}
		require.EqualValues(t, expected, ms.override)
	})
}

func TestMetricStore_PersistMetrics(t *testing.T) {
	WithMetricStore(func(ms *metricStore, _ *[]Query) {
		ms.Increment(NewSimpleKey(), "metric1.metric1.metric2")
		ms.Increment(NewSimpleKey(), "metric1")
		ms.Store(NewSimpleKey(), "metric3.metric4", 6)
		ms.Store(NewSimpleKey(), "metric2", 22)
		expectedEntries := map[storageKey]int{
			newStorageTestKey(ms, "metric1"): 3,
			newStorageTestKey(ms, "metric2"): 22,
			newStorageTestKey(ms, "metric3"): 6,
			newStorageTestKey(ms, "metric4"): 6,
		}
		_ = WithStore(ms, func(s *store.Store) bool {
			ms.persistMetrics(s)
			entries := make(map[storageKey]int)
			check.Must(s.RunForeachWithPrefix(PhonehomeKeyPrefix, func(keyb, valb []byte) (ok bool, err error) {
				var k storageKey
				var v int
				check.Must(k.Unmarshal(keyb))
				check.Must(json.Unmarshal(valb, &v))
				entries[k] = v
				return true, nil
			}))
			require.EqualValues(t, expectedEntries, entries)
			return true
		})
	})
}

func TestMetricStore_Send_Today(t *testing.T) {
	now := time.Now()
	cid := "cid"
	todayKey := newSimpleTestKey("map", now, cid)
	todayValue := 5
	WithMetricStore(func(ms *metricStore, sentQueries *[]Query) {
		// write the entry to database
		kb := check.MustValue(todayKey.Marshal())
		vb := check.MustValue(json.Marshal(todayValue))
		_ = WithStore(ms, func(s *store.Store) bool {
			check.Must(s.SetEntry(kb, vb))
			return true
		})
		// send the entries
		check.Must(ms.Send(context.Background()))
		// check that today's data is not sent and data exists in the database
		_, ft := findQuery(sentQueries, todayKey.Date(), cid)
		require.Equal(t, false, ft)
		keysToday := WithStore(ms, func(s *store.Store) [][]byte {
			return check.MustValue(s.GetKeysWithPrefix(datePrefix(todayKey.Date())))
		})
		require.Len(t, keysToday, 1)
	})
}

func TestMetricStore_Send_Yesterday(t *testing.T) {
	now := time.Now()
	cid := "cid"
	yesterdayCID := newSimpleTestKey("map", now.Add(-24*time.Hour), cid)
	yesterday := newSimpleTestKey("map", now.Add(-24*time.Hour), "")
	WithMetricStore(func(ms *metricStore, sentQueries *[]Query) {
		entries := map[storageKey]int{
			yesterdayCID: 10,
			yesterday:    20,
		}
		_ = WithStore(ms, func(s *store.Store) bool {
			for k, v := range entries {
				kb := check.MustValue(k.Marshal())
				vb := check.MustValue(json.Marshal(v))
				check.Must(s.SetEntry(kb, vb))
			}
			return true
		})
		check.Must(ms.Send(context.Background()))
		// check that yesterday's data is sent and data is deleted from the database
		queryWithCID, ok := findQuery(sentQueries, yesterday.Date(), cid)
		require.Equal(t, true, ok)
		query, ok := findQuery(sentQueries, yesterday.Date(), "")
		require.Equal(t, true, ok)
		keysYesterday := WithStore(ms, func(s *store.Store) [][]byte {
			return check.MustValue(s.GetKeysWithPrefix(datePrefix(yesterday.Date())))
		})
		require.Empty(t, keysYesterday)
		// check that queries are formed correctly
		require.Equal(t, 10, queryWithCID.MapRunCount)
		require.Equal(t, 20, query.MapRunCount)

	})
}

func findQuery(queries *[]Query, date, cid string) (Query, bool) {
	for _, q := range *queries {
		if q.Date == date && q.ClusterUUID == cid {
			return q, true
		}
	}
	return Query{}, false
}

func WithTempDir(fn func(string)) {
	dir, err := os.MkdirTemp("", "clc-metric-store-*")
	if err != nil {
		panic(fmt.Errorf("creating temp dir: %w", err))
	}
	defer func() {
		// errors are ignored
		os.RemoveAll(dir)
	}()
	fmt.Println(dir)
	fn(dir)
}

func WithMetricStore(fn func(ms *metricStore, queries *[]Query)) {
	WithTempDir(func(dir string) {
		queries := make([]Query, 0)
		sendQueriesFn := func(ctx context.Context, url string, q ...Query) error {
			queries = q
			return nil
		}
		ms := metricStore{
			inc:           make(map[storageKey]int),
			override:      make(map[storageKey]int),
			sa:            store.NewStoreAccessor(dir, log.NopLogger{}),
			sendQueriesFn: sendQueriesFn,
		}
		if err := ms.setGlobalMetrics(context.Background()); err != nil {
			panic(fmt.Errorf("setting global metrics: %w", err))
		}
		fn(&ms, &queries)
	})
}

func WithStore[T any](ms *metricStore, fn func(s *store.Store) T) T {
	val := check.MustValue(ms.sa.WithLock(func(s *store.Store) (any, error) {
		return fn(s), nil
	}))
	return val.(T)
}
