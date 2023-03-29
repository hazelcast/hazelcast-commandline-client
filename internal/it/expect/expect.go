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
}

func New(reader io.Reader) *Expect {
	e := &Expect{
		reader:    reader,
		bufReader: NewBuffer(100 * 1024),
		doneCh:    make(chan struct{}),
	}
	go e.read()
	return e
}

func (e *Expect) read() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-e.doneCh:
			return
		case <-ticker.C:
			check.MustValue(io.Copy(e.bufReader, e.reader))
		}
	}
}

func (e *Expect) Stop() {
	close(e.doneCh)
}

func (e *Expect) Reset() {
	e.bufReader.Reset()
}

func (e *Expect) String() string {
	return string(e.bufReader.Bytes())
}

func (e *Expect) Match(m Matcher, options ...Option) bool {
	e.Reset()
	o := Options{}
	for _, opt := range options {
		if err := opt(&o); err != nil {
			panic(fmt.Errorf("creating Match options: %w", err))
		}
	}
	ch := make(chan struct{})
	var done atomic.Bool
	go func() {
		for !done.Load() {
			if m.Match(e.String()) {
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
	mu   *sync.RWMutex
}

func NewBuffer(size int) *Buffer {
	return &Buffer{
		buf:  make([]byte, size),
		size: 0,
		mu:   &sync.RWMutex{},
	}
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	n = copy(b.buf[b.size:], p)
	b.size += n
	b.mu.Unlock()
	return n, nil
}

func (b *Buffer) Reset() {
	b.mu.Lock()
	b.size = 0
	b.mu.Unlock()
}

func (b *Buffer) Bytes() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.size == 0 {
		return nil
	}
	cp := make([]byte, b.size)
	copy(cp, b.buf[:b.size])
	return cp
}
