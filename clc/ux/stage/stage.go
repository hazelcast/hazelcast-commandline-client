package stage

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Statuser interface {
	SetProgress(progress float32)
	SetRemainingDuration(dur time.Duration)
}

type basicStatuser struct {
	text                 string
	textFmtWithRemaining string
	indexText            string
	sp                   clc.Spinner
}

func (s *basicStatuser) SetProgress(progress float32) {
	s.sp.SetProgress(progress)
}

func (s *basicStatuser) SetRemainingDuration(dur time.Duration) {
	text := s.text
	if dur > 0 {
		text = fmt.Sprintf(s.textFmtWithRemaining, dur)
	}
	s.sp.SetText(s.indexText + " " + text)
}

type Stage struct {
	ProgressMsg string
	SuccessMsg  string
	FailureMsg  string
	Func        func(status Statuser) error
}

type Provider internal.Iterator[Stage]

type Counter interface {
	StageCount() int
}

type FixedProvider struct {
	stages  []Stage
	offset  int
	current Stage
	err     error
}

func NewFixedProvider(stages ...Stage) *FixedProvider {
	return &FixedProvider{stages: stages}
}

func (sp *FixedProvider) Next() bool {
	if sp.offset >= len(sp.stages) {
		return false
	}
	sp.current = sp.stages[sp.offset]
	sp.offset++
	return true
}

func (sp *FixedProvider) Value() Stage {
	return sp.current
}

func (sp *FixedProvider) Err() error {
	return sp.err
}

func (sp *FixedProvider) StageCount() int {
	return len(sp.stages)
}

func Execute(ctx context.Context, ec plug.ExecContext, sp Provider) error {
	ss := &basicStatuser{}
	var index int
	var stageCount int
	if sc, ok := sp.(Counter); ok {
		stageCount = sc.StageCount()
	}
	for sp.Next() {
		if sp.Err() != nil {
			return sp.Err()
		}
		stg := sp.Value()
		index++
		ss.text = stg.ProgressMsg
		ss.textFmtWithRemaining = stg.ProgressMsg + " (%s left)"
		if stageCount > 0 {
			d := paddedIntFormat(stageCount)
			ss.indexText = fmt.Sprintf("["+d+"/%d]", index, stageCount)
		} else {
			ss.indexText = ""
		}
		_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, spinner clc.Spinner) (any, error) {
			ss.sp = spinner
			ss.SetRemainingDuration(0)
			return nil, stg.Func(ss)
		})
		if err != nil {
			ec.PrintlnUnnecessary(fmt.Sprintf("FAIL %s: %s", stg.FailureMsg, err.Error()))
			return err
		}
		stop()
		ec.PrintlnUnnecessary(fmt.Sprintf("OK %s %s.", ss.indexText, stg.SuccessMsg))
	}
	return nil
}

// paddedIntFormat returns the fmt string that can fit the given integer.
// The padding contains spaces.
func paddedIntFormat(maxValue int) string {
	d := int(math.Ceil(math.Log10(float64(maxValue))))
	return "%" + strconv.Itoa(d) + "d"
}
