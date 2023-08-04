package store

import (
	"errors"
	"sync"

	badger "github.com/dgraph-io/badger/v4"
)

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrDatabaseNotOpen = errors.New("database is not open")
)

type StoreAccessor struct {
	store *Store
	mu    sync.Mutex
}

func NewStoreAccessor(dir string) *StoreAccessor {
	return &StoreAccessor{
		store: newStore(dir),
	}
}

func (s *StoreAccessor) WithLock(fn func(s *Store) (any, error)) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.store.open(); err != nil {
		return nil, err
	}
	defer s.store.close()
	return fn(s.store)
}

type Store struct {
	dir string
	db  *badger.DB
}

func newStore(dir string) *Store {
	return &Store{
		dir: dir,
	}
}

func (b *Store) open() error {
	opts := badger.DefaultOptions(b.dir)
	opts.Logger = noplogger{}
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	b.db = db
	return nil
}

type noplogger struct{}

func (noplogger) Errorf(string, ...interface{})   {}
func (noplogger) Warningf(string, ...interface{}) {}
func (noplogger) Infof(string, ...interface{})    {}
func (noplogger) Debugf(string, ...interface{})   {}

func (b *Store) close() error {
	return b.db.Close()
}

func (b *Store) SetEntry(key, val []byte) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, val)
		return txn.SetEntry(e)
	})
	return err
}

func (b *Store) GetEntry(key []byte) ([]byte, error) {
	var item []byte
	err := b.db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrKeyNotFound
			}
			return err
		}
		item, err = it.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (b *Store) GetKeysWithPrefix(prefix string) ([][]byte, error) {
	keys := [][]byte{}
	pb := []byte(prefix)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(pb); it.ValidForPrefix(pb); it.Next() {
			keyCopy := it.Item().KeyCopy(nil)
			keys = append(keys, keyCopy)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

type updateFn = func(current []byte, found bool) []byte

func (b *Store) UpdateEntry(key []byte, f updateFn) error {
	var item []byte
	err := b.db.Update(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		found := true
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				found = false
			} else {
				return err
			}
		} else {
			item, err = it.ValueCopy(nil)
			if err != nil {
				return err
			}
		}
		newVal := f(item, found)
		return txn.Set(key, newVal)
	})
	return err
}

type foreachFn = func(key, val []byte)

func (b *Store) RunForeachWithPrefix(prefix string, f foreachFn) error {
	p := []byte(prefix)
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			f(k, v)
		}
		return nil
	})
	return err
}

func (b *Store) DeleteEntriesWithPrefix(prefix string) error {
	keys, err := b.GetKeysWithPrefix(prefix)
	if err != nil {
		return err
	}
	err = b.db.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
