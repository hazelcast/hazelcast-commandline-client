package project

import (
	"context"
	"os"
	"testing"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestCreateCommand(t *testing.T) {
	os.Setenv(envTemplateSource, "https://github.com/kutluhanmetin")
	testCases := []struct {
		inputTemplateName string
		inputOutputDir    string
		inputArgs         []string
		testProjectDir    string
	}{
		{
			inputTemplateName: "simple-streaming-pipeline-template",
			inputOutputDir:    "my-simple-streaming-pipeline",
			inputArgs:         []string{"rootProjectName=simple-streaming-pipeline"},
			testProjectDir:    "../../../examples/platform/simple-streaming-pipeline",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.inputTemplateName, func(t *testing.T) {
			defer teardown(tc.inputOutputDir)
			tcx := it.TestContext{T: t}
			tcx.Tester(func(tcx it.TestContext) {
				ctx := context.Background()
				tcx.WithReset(func() {
					cmd := []string{"project", "create", tc.inputTemplateName, tc.inputOutputDir}
					cmd = append(cmd, tc.inputArgs...)
					check.Must(tcx.CLC().Execute(ctx, cmd...))
				})
				tcx.WithReset(func() {
					check.Must(compareDirectories(tc.inputOutputDir, tc.testProjectDir))
				})
			})
		})
	}
}

func teardown(dir string) {
	os.RemoveAll(dir)
}
