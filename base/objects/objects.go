package objects

import (
	"context"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func GetAll(ctx context.Context, ec plug.ExecContext, typeFilter string, showHidden bool) ([]types.DistributedObjectInfo, error) {
	objs, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText("Getting distributed objects")
		return ci.Client().GetDistributedObjectsInfo(ctx)
	})
	if err != nil {
		return nil, err
	}
	stop()
	var r []types.DistributedObjectInfo
	typeFilter = strings.ToLower(typeFilter)
	for _, o := range objs.([]types.DistributedObjectInfo) {
		if !showHidden && (o.Name == "" || strings.HasPrefix(o.Name, "__")) {
			continue
		}
		if o.Name == "" {
			o.Name = "(no name)"
		}
		if typeFilter == "" {
			r = append(r, o)
			continue
		}
		if typeFilter == ShortType(o.ServiceName) {
			r = append(r, o)
		}
	}
	sort.Slice(r, func(i, j int) bool {
		// first sort by type, then name
		ri := r[i]
		rj := r[j]
		if ri.ServiceName < rj.ServiceName {
			return true
		}
		if ri.ServiceName > rj.ServiceName {
			return false
		}
		return ri.Name < rj.Name
	})
	return r, nil
}

func ShortType(svcName string) string {
	s := strings.TrimSuffix(strings.TrimPrefix(svcName, "hz:impl:"), "Service")
	return strings.ToLower(s)
}
