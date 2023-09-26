package stage_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestStage(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "execute", f: executeTest},
		{name: "execute_WithFailureTest", f: execute_WithFailureTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func executeTest(t *testing.T) {
	stages := []stage.Stage[any]{
		{
			ProgressMsg: "Progressing 1",
			SuccessMsg:  "Success 1",
			FailureMsg:  "Failure 1",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				time.Sleep(1 * time.Millisecond)
				return nil, nil
			},
		},
		{
			ProgressMsg: "Progressing 2",
			SuccessMsg:  "Success 2",
			FailureMsg:  "Failure 2",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				for i := 0; i < 5; i++ {
					status.SetProgress(float32(i+1) / float32(5))
				}
				time.Sleep(1 * time.Millisecond)
				return nil, nil
			},
		},
		{
			ProgressMsg: "Progressing 3",
			SuccessMsg:  "Success 3",
			FailureMsg:  "Failure 3",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				status.SetText("Custom text")
				for i := 0; i < 5; i++ {
					status.SetRemainingDuration(5*time.Second - time.Duration(i+1)*time.Second)
				}
				time.Sleep(1 * time.Millisecond)
				return nil, nil
			},
		},
	}
	ec := it.NewExecuteContext(nil)
	_, err := stage.Execute[any](context.TODO(), ec, nil, stage.NewFixedProvider(stages...))
	assert.NoError(t, err)
	texts := []string{
		" [1/3] Progressing 1",
		" [2/3] Progressing 2",
		" [3/3] Progressing 3",
		" [3/3] Custom text",
		" [3/3] Custom text (4s left)",
		" [3/3] Custom text (3s left)",
		" [3/3] Custom text (2s left)",
		" [3/3] Custom text (1s left)",
		" [3/3] Custom text",
	}
	assert.Equal(t, texts, ec.Spinner.Texts)
	progresses := []float32{0.2, 0.4, 0.6, 0.8, 1}
	assert.Equal(t, progresses, ec.Spinner.Progresses)
	text := "OK [1/3] Success 1.\nOK [2/3] Success 2.\nOK [3/3] Success 3.\n"
	assert.Equal(t, text, ec.StdoutText())
}

func execute_WithFailureTest(t *testing.T) {
	stages := []stage.Stage[any]{
		{
			ProgressMsg: "Progressing 1",
			SuccessMsg:  "Success 1",
			FailureMsg:  "Failure 1",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				return nil, fmt.Errorf("some error")
			},
		},
		{
			ProgressMsg: "Progressing 2",
			SuccessMsg:  "Success 2",
			FailureMsg:  "Failure 2",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				return nil, nil
			},
		},
	}
	ec := it.NewExecuteContext(nil)
	_, err := stage.Execute[any](context.TODO(), ec, nil, stage.NewFixedProvider(stages...))
	assert.Error(t, err)
	texts := []string{" [1/2] Progressing 1"}
	assert.Equal(t, texts, ec.Spinner.Texts)
	text := "ERROR Failure 1: some error\n"
	assert.Equal(t, text, ec.StdoutText())
}
