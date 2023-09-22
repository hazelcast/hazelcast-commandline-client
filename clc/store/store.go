package store

import (
	"errors"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type StoreAccessor struct {
	store *Store
	mu    sync.Mutex
}

func NewStoreAccessor(dir string, logger log.Logger) *StoreAccessor {
	return &StoreAccessor{
		store: newStore(dir, logger),
	}
}

func (sa *StoreAccessor) WithLock(fn func(s *Store) (any, error)) (any, error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	if err := sa.store.open(); err != nil {
		return nil, err
	}
	defer sa.store.close()
	return fn(sa.store)
}

type Store struct {
	dir    string
	db     *badger.DB
	logger log.Logger
}

func newStore(dir string, logger log.Logger) *Store {
	return &Store{
		dir:    dir,
		logger: logger,
	}
}

func (s *Store) open() error {
	opts := badger.DefaultOptions(s.dir)
	opts.Logger = badgerLogger{s.logger}
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *Store) close() error {
	return s.db.Close()
}

type badgerLogger struct {
	logger log.Logger
}

func (l badgerLogger) Errorf(log string, args ...any) {
	log = "Store ERROR: " + log
	l.logger.Debugf(log, args...)
}
func (l badgerLogger) Warningf(log string, args ...any) {
	log = "Store WARNING: " + log
	l.logger.Debugf(log, args...)
}
func (l badgerLogger) Infof(log string, args ...any) {
	log = "Store INFO: " + log
	l.logger.Debugf(log, args...)
}
func (l badgerLogger) Debugf(log string, args ...any) {
	log = "Store DEBUG: " + log
	l.logger.Debugf(log, args...)
}

type Option interface {
	ModifyEntry(*badger.Entry)
}

type OptionWithTTL time.Duration

func (s OptionWithTTL) ModifyEntry(e *badger.Entry) {
	e.WithTTL(time.Duration(s))
}

func (s *Store) SetEntry(key, val []byte, opts ...Option) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, val)
		for _, opt := range opts {
			opt.ModifyEntry(e)
		}
		return txn.SetEntry(e)
	})
	return err
}

func (s *Store) GetEntry(key []byte) ([]byte, error) {
	var item []byte
	err := s.db.View(func(txn *badger.Txn) error {
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

func (s *Store) GetKeysWithPrefix(prefix string) ([][]byte, error) {
	keys := [][]byte{}
	pb := []byte(prefix)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(pb); it.ValidForPrefix(pb); it.Next() {
			c := it.Item().KeyCopy(nil)
			keys = append(keys, c)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

type UpdateFunc = func(current []byte, found bool) []byte

func (s *Store) UpdateEntry(key []byte, f UpdateFunc, opts ...Option) error {
	var item []byte
	err := s.db.Update(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		found := true
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
			found = false
		} else {
			item, err = it.ValueCopy(nil)
			if err != nil {
				return err
			}
		}
		newVal := f(item, found)
		e := badger.NewEntry(key, newVal)
		for _, opt := range opts {
			opt.ModifyEntry(e)
		}
		return txn.Set(key, newVal)
	})
	return err
}

type ForeachFunc = func(key, val []byte) (ok bool, err error)

func (s *Store) RunForeachWithPrefix(prefix string, f ForeachFunc) error {
	p := []byte(prefix)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			item := it.Item()
			// Note: We always copy key and value, if there is a performance issue this might be a reason
			k := item.KeyCopy(nil)
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			ok, err := f(k, v)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
		}
		return nil
	})
	return err
}

func (s *Store) DeleteEntriesWithPrefixes(prefixes ...string) error {
	prefixesb := [][]byte{}
	for _, p := range prefixes {
		prefixesb = append(prefixesb, []byte(p))
	}
	err := s.db.DropPrefix(prefixesb...)
	return err
}
