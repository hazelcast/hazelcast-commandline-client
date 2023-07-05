package ux

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

type StageStatuser interface {
	SetProgress(progress float32)
	SetRemainingDuration(dur time.Duration)
}

type stageStatuser struct {
	text                 string
	textFmtWithRemaining string
	indexText            string
	sp                   clc.Spinner
}

func (s *stageStatuser) SetProgress(progress float32) {
	s.sp.SetProgress(progress)
}

func (s *stageStatuser) SetRemainingDuration(dur time.Duration) {
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
	Func        func(status StageStatuser) error
}

type StageProvider internal.Iterator[Stage]

type StageCounter interface {
	StageCount() int
}

type FixedStageProvider struct {
	stages  []Stage
	offset  int
	current Stage
	err     error
}

func NewFixedStageProvider(stages ...Stage) *FixedStageProvider {
	return &FixedStageProvider{stages: stages}
}

func (sp *FixedStageProvider) Next() bool {
	if sp.offset >= len(sp.stages) {
		return false
	}
	sp.current = sp.stages[sp.offset]
	sp.offset++
	return true
}

func (sp *FixedStageProvider) Value() Stage {
	return sp.current
}

func (sp *FixedStageProvider) Err() error {
	return sp.err
}

func (sp *FixedStageProvider) StageCount() int {
	return len(sp.stages)
}

func ExecuteStages(ctx context.Context, ec plug.ExecContext, sp StageProvider) error {
	ss := &stageStatuser{}
	var index int
	var stageCount int
	if sc, ok := sp.(StageCounter); ok {
		stageCount = sc.StageCount()
	}
	for sp.Next() {
		if sp.Err() != nil {
			return sp.Err()
		}
		stage := sp.Value()
		index++
		ss.text = stage.ProgressMsg
		ss.textFmtWithRemaining = stage.ProgressMsg + " (%s left)"
		if stageCount > 0 {
			digits := int(math.Ceil(math.Log10(float64(stageCount)))) + 1
			d := "%" + strconv.Itoa(digits) + "d"
			ss.indexText = fmt.Sprintf("["+d+"/%d]", index, stageCount)
		} else {
			ss.indexText = ""
		}
		_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, spinner clc.Spinner) (any, error) {
			ss.sp = spinner
			ss.SetRemainingDuration(0)
			return nil, stage.Func(ss)
		})
		if err != nil {
			ec.PrintlnUnnecessary(fmt.Sprintf("FAIL %s: %s", stage.FailureMsg, err.Error()))
			return err
		}
		stop()
		ec.PrintlnUnnecessary(fmt.Sprintf("OK %s %s.", ss.indexText, stage.SuccessMsg))
	}
	return nil
}
