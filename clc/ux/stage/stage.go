package stage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	hzerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
)

type Statuser[T any] interface {
	SetProgress(progress float32)
	SetRemainingDuration(dur time.Duration)
	Value() T
}

type basicStatuser[T any] struct {
	text                 string
	textFmtWithRemaining string
	indexText            string
	sp                   clc.Spinner
	value                T
}

func (s basicStatuser[T]) SetProgress(progress float32) {
	s.sp.SetProgress(progress)
}

func (s basicStatuser[T]) SetRemainingDuration(dur time.Duration) {
	text := s.text
	if dur > 0 {
		text = fmt.Sprintf(s.textFmtWithRemaining, dur)
	}
	s.sp.SetText(s.indexText + " " + text)
}

func (s basicStatuser[T]) Value() T {
	return s.value
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
		ss := basicStatuser[T]{value: value}
		ss.text = stg.ProgressMsg
		ss.textFmtWithRemaining = stg.ProgressMsg + " (%s left)"
		ss.indexText = ""
		if stageCount > 1 {
			d := str.SpacePaddedIntFormat(stageCount)
			ss.indexText = fmt.Sprintf("["+d+"/%d]", index, stageCount)
		}
		v, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, spinner clc.Spinner) (any, error) {
			ss.sp = spinner
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
				ec.PrintlnUnnecessary(fmt.Sprintf("ERROR %s %s: %s", ss.indexText, stg.FailureMsg, ie.Unwrap().Error()))
			} else {
				ec.PrintlnUnnecessary(fmt.Sprintf("ERROR %s: %s", stg.FailureMsg, err.Error()))
				return value, hzerrors.WrappedError{Err: err}
			}
		}
		stop()
		if err == nil {
			ec.PrintlnUnnecessary(fmt.Sprintf("OK %s %s.", ss.indexText, stg.SuccessMsg))
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
