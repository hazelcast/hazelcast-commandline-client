package _cluster_test

import (
	"context"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestCluster(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "Get_Noninteractive", f: get_NonInteractiveTest},
		{name: "ListMembers_NonInteractive", f: listMembers_NonInteractiveTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func get_NonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "cluster", "get"))
			tcx.AssertStdoutContains(tcx.ClientConfig.Cluster.Name)
			ci := hazelcast.NewClientInternal(tcx.Client)
			tcx.AssertStdoutContains(ci.ClusterID().String())
		})
	})
}

func listMembers_NonInteractiveTest(t *testing.T) {
	tcx := it.TestContext{T: t}
	tcx.Tester(func(tcx it.TestContext) {
		ctx := context.Background()
		tcx.WithReset(func() {
			check.Must(tcx.CLC().Execute(ctx, "cluster", "list-members"))
			ci := hazelcast.NewClientInternal(tcx.Client)
			for _, mem := range ci.OrderedMembers() {
				tcx.AssertStdoutContains(mem.UUID.String())
			}
		})
	})
}

func objectExists(sn, name string, objects []types.DistributedObjectInfo) bool {
	for _, obj := range objects {
		if sn == obj.ServiceName && name == obj.Name {
			return true
		}
	}
	return false
}
