package stage_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestExecute(t *testing.T) {
	stages := []stage.Stage{
		{
			ProgressMsg: "Progressing 1",
			SuccessMsg:  "Success 1",
			FailureMsg:  "Failure 1",
			Func: func(status stage.Statuser) error {
				time.Sleep(1 * time.Millisecond)
				return nil
			},
		},
		{
			ProgressMsg: "Progressing 2",
			SuccessMsg:  "Success 2",
			FailureMsg:  "Failure 2",
			Func: func(status stage.Statuser) error {
				for i := 0; i < 5; i++ {
					status.SetProgress(float32(i+1) / float32(5))
				}
				time.Sleep(1 * time.Millisecond)
				return nil
			},
		},
		{
			ProgressMsg: "Progressing 3",
			SuccessMsg:  "Success 3",
			FailureMsg:  "Failure 3",
			Func: func(status stage.Statuser) error {
				for i := 0; i < 5; i++ {
					status.SetRemainingDuration(5*time.Second - time.Duration(i+1)*time.Second)
				}
				time.Sleep(1 * time.Millisecond)
				return nil
			},
		},
	}
	ec := it.NewExecuteContext(nil)
	err := stage.Execute(context.TODO(), ec, stage.NewFixedProvider(stages...))
	assert.NoError(t, err)
	texts := []string{
		"[1/3] Progressing 1",
		"[2/3] Progressing 2",
		"[3/3] Progressing 3",
		"[3/3] Progressing 3 (4s left)",
		"[3/3] Progressing 3 (3s left)",
		"[3/3] Progressing 3 (2s left)",
		"[3/3] Progressing 3 (1s left)",
		"[3/3] Progressing 3",
	}
	assert.Equal(t, texts, ec.Spinner.Texts)
	progresses := []float32{0.2, 0.4, 0.6, 0.8, 1}
	assert.Equal(t, progresses, ec.Spinner.Progresses)
}
