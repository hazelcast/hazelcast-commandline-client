package cluster

import (
	"context"
	"math/rand"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/types"
)

func RandomMember(ctx context.Context, ci *hazelcast.ClientInternal) (types.UUID, error) {
	var mi cluster.MemberInfo
	for {
		if ctx.Err() != nil {
			return types.UUID{}, ctx.Err()
		}
		mems := ci.OrderedMembers()
		if len(mems) != 0 {
			mi = mems[rand.Intn(len(mems))]
			if ci.ConnectedToMember(mi.UUID) {
				return mi.UUID, nil
			}
		}
	}
}
