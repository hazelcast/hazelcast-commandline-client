package project

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestCreateCommand(t *testing.T) {
	testCases := []struct {
		inputTemplateName string
		inputOutputDir    string
		inputArgs         []string
		testProjectDir    string
	}{
		{
			inputTemplateName: "my-template",
			inputOutputDir:    "my-project",
			inputArgs:         []string{"myName=foo", "mySurname=bar"},
			testProjectDir:    "testdata/my-project",
		},
		{
			inputTemplateName: "simple-streaming-pipeline-template",
			inputOutputDir:    "my-simple-streaming-pipeline",
			testProjectDir:    "../../../examples/platform/simple-streaming-pipeline",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.inputTemplateName, func(t *testing.T) {
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				ctx := context.Background()
				tcx.WithReset(func() {
					cmd := []string{"project", "create", tc.inputTemplateName, fmt.Sprintf("--%s", projectOutputDir), tc.inputOutputDir}
					cmd = append(cmd, tc.inputArgs...)
					check.Must(tcx.CLC().Execute(ctx, cmd...))
					check.Must(compareDirectories(tc.inputOutputDir, tc.testProjectDir))
				})
			})
		})
		teardown(tc.inputOutputDir)
	}
}

func teardown(dir string) {
	os.RemoveAll(dir)
}
