package stage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
)

type Statuser[T any] interface {
	SetText(text string)
	SetProgress(progress float32)
	SetRemainingDuration(dur time.Duration)
	Value() T
}

type basicStatuser[T any] struct {
	sp    clc.Spinner
	value T
	mu    *sync.RWMutex
	index int
	count int
	text  string
}

func newBasicStatuser[T any](value T) *basicStatuser[T] {
	return &basicStatuser[T]{
		value: value,
		mu:    &sync.RWMutex{},
	}
}

func (s *basicStatuser[T]) SetText(text string) {
	s.mu.Lock()
	s.text = text
	sp := s.sp
	s.mu.Unlock()
	if sp != nil {
		s.sp.SetText(" " + s.IndexText() + " " + text)
	}
}

func (s *basicStatuser[T]) SetProgress(progress float32) {
	s.sp.SetProgress(progress)
}

func (s *basicStatuser[T]) SetRemainingDuration(dur time.Duration) {
	s.mu.RLock()
	text := s.text
	if dur > 0 {
		text = fmt.Sprintf(s.textFmtWithRemaining(), dur)
	}
	s.mu.RUnlock()
	s.sp.SetText(" " + s.IndexText() + " " + text)
}

func (s *basicStatuser[T]) Value() T {
	return s.value
}

func (s *basicStatuser[T]) SetIndex(index, count int) {
	s.mu.Lock()
	s.index = index
	s.count = count
	s.mu.Unlock()
}

func (s *basicStatuser[T]) IndexText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.index == 0 || s.count < 2 {
		return ""
	}
	d := str.SpacePaddedIntFormat(s.count)
	return fmt.Sprintf("["+d+"/%d]", s.index, s.count)
}

func (s *basicStatuser[T]) SetSpinner(sp clc.Spinner) {
	s.mu.Lock()
	s.sp = sp
	s.mu.Unlock()
}

func (s *basicStatuser[T]) textFmtWithRemaining() string {
	return s.text + " (%s left)"
}

type Stage[T any] struct {
	ProgressMsg string
	SuccessMsg  string
	FailureMsg  string
	Func        func(ctx context.Context, status Statuser[T]) (T, error)
}

type Provider[T any] internal.Iterator[Stage[T]]

type Counter interface {
	StageCount() int
}

type FixedProvider[T any] struct {
	stages  []Stage[T]
	offset  int
	current Stage[T]
	err     error
}

func NewFixedProvider[T any](stages ...Stage[T]) *FixedProvider[T] {
	return &FixedProvider[T]{stages: stages}
}

func (sp *FixedProvider[T]) Next() bool {
	if sp.offset >= len(sp.stages) {
		return false
	}
	sp.current = sp.stages[sp.offset]
	sp.offset++
	return true
}

func (sp *FixedProvider[T]) Value() Stage[T] {
	return sp.current
}

func (sp *FixedProvider[T]) Err() error {
	return sp.err
}

func (sp *FixedProvider[T]) StageCount() int {
	return len(sp.stages)
}

func Execute[T any](ctx context.Context, ec plug.ExecContext, value T, sp Provider[T]) (T, error) {
	var index int
	var stageCount int
	if sc, ok := sp.(Counter); ok {
		stageCount = sc.StageCount()
	}
	for sp.Next() {
		if sp.Err() != nil {
			return value, sp.Err()
		}
		stg := sp.Value()
		index++
		ss := newBasicStatuser(value)
		ss.SetText(stg.ProgressMsg)
		ss.SetIndex(index, stageCount)
		v, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, spinner clc.Spinner) (any, error) {
			ss.SetSpinner(spinner)
			ss.SetRemainingDuration(0)
			v, err := stg.Func(ctx, ss)
			if err != nil {
				return nil, err
			}
			return any(v), nil
		})
		if err != nil {
			var ie ignoreError
			if errors.As(err, &ie) {
				// the error can be ignored
				ec.PrintlnUnnecessary(fmt.Sprintf("ERROR %s %s: %s", ss.IndexText(), stg.FailureMsg, ie.Unwrap().Error()))
			} else {
				return value, fmt.Errorf("%s: %w", stg.FailureMsg, err)
			}
		}
		stop()
		if err == nil {
			ec.PrintlnUnnecessary(fmt.Sprintf("OK %s %s.", ss.IndexText(), stg.SuccessMsg))
		}
		if v == nil {
			var vv T
			value = vv
		} else {
			value = v.(T)
		}
	}
	return value, nil
}
