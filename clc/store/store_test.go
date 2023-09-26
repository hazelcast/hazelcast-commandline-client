package store

import (
	"fmt"
	"os"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

func TestStore_GetSetEntry(t *testing.T) {
	withStore(func(s *Store) {
		check.Must(insertValues(s.db, map[string][]byte{
			"key1": []byte("val"),
		}))
		valb := check.MustValue(s.GetEntry([]byte("key1")))
		require.Equal(t, []byte("val"), valb)
		check.Must(s.SetEntry([]byte("key1"), []byte("valnew")))
		valnew := check.MustValue(s.GetEntry([]byte("key1")))
		require.Equal(t, []byte("valnew"), valnew)
	})
}

func TestStore_SettEntryWithTTL(t *testing.T) {
	withStore(func(s *Store) {
		check.Must(s.SetEntry([]byte("key1"), []byte("val1")))
		check.Must(s.SetEntry([]byte("key2"), []byte("val2"), OptionWithTTL(100*time.Millisecond)))
		check.Must(s.SetEntry([]byte("key3"), []byte("val3"), OptionWithTTL(5*time.Second)))
		time.Sleep(200 * time.Millisecond)
		entries := check.MustValue(getAllEntries(s.db))
		require.Equal(t, entries["key1"], []byte("val1"))
		require.NotContains(t, entries, "key2")
		require.Equal(t, entries["key3"], []byte("val3"))
	})
}

func TestStore_GetKeysWithPrefix(t *testing.T) {
	withStore(func(s *Store) {
		check.Must(insertValues(s.db, map[string][]byte{
			"prefix.key1": []byte(""),
			"prefix.key2": []byte(""),
			"noprefix":    []byte(""),
		}))
		vals := check.MustValue(s.GetKeysWithPrefix("prefix"))
		expected := [][]byte{[]byte("prefix.key1"), []byte("prefix.key2")}
		require.ElementsMatch(t, expected, vals)

	})
}

func TestStore_UpdateEntry(t *testing.T) {
	withStore(func(s *Store) {
		check.Must(s.UpdateEntry([]byte("key"), func(current []byte, found bool) []byte {
			if !found {
				return []byte("notexist")
			}
			return nil
		}))
		valb := check.MustValue(s.GetEntry([]byte("key")))
		require.Equal(t, []byte("notexist"), valb)
		check.Must(s.UpdateEntry([]byte("key"), func(current []byte, found bool) []byte {
			if found {
				return append(current, []byte(".nowexist")...)
			}
			return nil
		}))
		valnew := check.MustValue(s.GetEntry([]byte("key")))
		require.Equal(t, []byte("notexist.nowexist"), valnew)
	})
}

func TestStore_RunForeachWithPrefix(t *testing.T) {
	fromStore := make(map[string][]byte)
	withStore(func(s *Store) {
		check.Must(insertValues(s.db, map[string][]byte{
			"prefix.key1": []byte(""),
			"prefix.key2": []byte(""),
			"noprefix":    []byte(""),
		}))
		check.Must(s.RunForeachWithPrefix("prefix", func(key, val []byte) (bool, error) {
			fromStore[string(key)] = val
			return true, nil
		}))
		expected := map[string][]byte{
			"prefix.key1": nil,
			"prefix.key2": nil,
		}
		require.EqualValues(t, expected, fromStore)
	})
}

func TestStore_DeleteEntriesWithPrefix(t *testing.T) {
	withStore(func(s *Store) {
		check.Must(insertValues(s.db, map[string][]byte{
			"1.prefix.key1": []byte(""),
			"1.prefix.key2": []byte(""),
			"2.prefix.key3": []byte(""),
			"noprefix":      []byte(""),
		}))
		check.Must(s.DeleteEntriesWithPrefixes("1.prefix", "2.prefix"))
		entries := check.MustValue(getAllEntries(s.db))
		expected := map[string][]byte{
			"noprefix": nil,
		}
		require.EqualValues(t, expected, entries)
	})
}

func getAllEntries(db *badger.DB) (map[string][]byte, error) {
	m := make(map[string][]byte)
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			b, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			m[string(k)] = b
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func insertValues(db *badger.DB, vals map[string][]byte) error {
	err := db.Update(func(txn *badger.Txn) error {
		for k, v := range vals {
			err := txn.SetEntry(badger.NewEntry([]byte(k), v))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func WithTempDir(fn func(dir string)) {
	dir, err := os.MkdirTemp("", "clc-store-*")
	if err != nil {
		panic(fmt.Errorf("creating temp dir: %w", err))
	}
	defer func() {
		// errors are ignored
		os.RemoveAll(dir)
	}()
	fn(dir)
}

func withStore(fn func(s *Store)) {
	WithTempDir(func(dir string) {
		s := NewStoreAccessor(dir, log.NopLogger{})
		s.WithLock(func(s *Store) (any, error) {
			fn(s)
			return nil, nil
		})
	})
}
