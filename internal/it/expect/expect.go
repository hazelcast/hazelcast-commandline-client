package expect

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type Expect struct {
	reader    io.Reader
	bufReader *Buffer
	doneCh    chan struct{}
	mu        *sync.RWMutex
}

func New(reader io.Reader) *Expect {
	e := &Expect{
		reader:    reader,
		bufReader: NewBuffer(100 * 1024),
		doneCh:    make(chan struct{}),
		mu:        &sync.RWMutex{},
	}
	go e.read()
	return e
}

func (e *Expect) read() {
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-e.doneCh:
			return
		case <-ticker.C:
			e.mu.Lock()
			check.MustValue(io.Copy(e.bufReader, e.reader))
			e.mu.Unlock()
		}
	}
}

func (e *Expect) Stop() {
	close(e.doneCh)
}

func (e *Expect) Reset() {
	e.mu.Lock()
	e.bufReader.Reset()
	e.mu.Unlock()
}

func (e *Expect) String() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return string(e.bufReader.Bytes())
}

func (e *Expect) Match(m Matcher, options ...Option) bool {
	o := Options{}
	for _, opt := range options {
		if err := opt(&o); err != nil {
			panic(fmt.Errorf("creating Match options: %w", err))
		}
	}
	if o.delay > 0 {
		time.Sleep(o.delay)
	}
	ch := make(chan struct{})
	var done atomic.Bool
	go func() {
		for !done.Load() {
			e.mu.RLock()
			ok := m.Match(e.String())
			e.mu.RUnlock()
			if ok {
				ch <- struct{}{}
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	timeout := time.Duration(1<<63 - 1)
	if o.timeout > 0 {
		timeout = o.timeout
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-e.doneCh:
		break
	case <-timer.C:
		break
	case <-ch:
		return true
	}
	done.Store(true)
	return false
}

type Buffer struct {
	buf  []byte
	size int
}

func NewBuffer(size int) *Buffer {
	return &Buffer{
		buf: make([]byte, size),
	}
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	n = copy(b.buf[b.size:], p)
	b.size += n
	return n, nil
}

func (b *Buffer) Reset() {
	b.size = 0
}

func (b *Buffer) Bytes() []byte {
	if b.size == 0 {
		return nil
	}
	cp := make([]byte, b.size)
	copy(cp, b.buf[:b.size])
	return cp
}
