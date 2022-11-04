package objects

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/types"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	Map              = "map"
	ReplicatedMap    = "replicatedMap"
	MultiMap         = "multiMap"
	Queue            = "queue"
	Topic            = "topic"
	List             = "list"
	Set              = "set"
	PNCounter        = "PNCounter"
	FlakeIDGenerator = "flakeIdGenerator"
	// unsupported types by go client
	Cache                = "cache"
	EventJournal         = "eventJournal"
	Ringbuffer           = "ringBuffer"
	FencedLock           = "fencedLock"
	ISemaphore           = "semaphore"
	IAtomicLong          = "atomicLong"
	IAtomicReference     = "atomicReference"
	ICountdownLatch      = "countdownLatch"
	CardinalityEstimator = "cardinalityEstimator"
)

var objTypes = []string{
	Map,
	ReplicatedMap,
	MultiMap,
	Queue,
	Topic,
	List,
	Set,
	PNCounter,
	FlakeIDGenerator,
	// unsupported types by go client
	Cache,
	EventJournal,
	Ringbuffer,
	FencedLock,
	ISemaphore,
	IAtomicLong,
	IAtomicReference,
	ICountdownLatch,
	CardinalityEstimator,
}

const (
	flagType       = "type"
	flagShowHidden = "show-hidden"
)

type ObjectsListCommand struct{}

func (cm ObjectsListCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list [object-type]")
	cc.SetCommandHelp("list distributed objects, optionally filter by type", "list distributed objects")
	typeHelp := fmt.Sprintf("filter by type, one of: %s", strings.Join(objTypes, ","))
	cc.AddStringFlag(flagType, "t", "", false, typeHelp)
	cc.AddBoolFlag(flagShowHidden, "", false, false, "show hidden and system objects")
	return nil
}

func (cm ObjectsListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	typeFilter := ec.Props().GetString(flagType)
	showHidden := ec.Props().GetBool(flagShowHidden)
	objs, err := getObjects(ctx, ec, typeFilter, showHidden)
	if err != nil {
		return err
	}
	for _, o := range objs {
		valueCol := output.Column{
			Name:  "Object Name",
			Type:  serialization.TypeString,
			Value: o.Name,
		}
		if typeFilter != "" {
			ec.AddOutputRows(output.Row{valueCol})
			continue
		}
		ec.AddOutputRows(output.Row{
			output.Column{
				Name:  "Service Name",
				Type:  serialization.TypeString,
				Value: shortType(o.ServiceName),
			},
			valueCol,
		})
	}
	return nil
}

func getObjects(ctx context.Context, ec plug.ExecContext, typeFilter string, showHidden bool) ([]types.DistributedObjectInfo, error) {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	objs, err := ec.ExecuteBlocking(ctx, "Getting distributed objects", func(ctx context.Context) (any, error) {
		return ci.Client().GetDistributedObjectsInfo(ctx)
	})
	if err != nil {
		return nil, err
	}
	var r []types.DistributedObjectInfo
	for _, o := range objs.([]types.DistributedObjectInfo) {
		if !showHidden && strings.HasPrefix(o.Name, "__") {
			continue
		}
		if typeFilter == "" {
			r = append(r, o)
			continue
		}
		if typeFilter == shortType(o.ServiceName) {
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

func shortType(svcName string) string {
	return strings.TrimSuffix(strings.TrimPrefix(svcName, "hz:impl:"), "Service")
}

func init() {
	Must(plug.Registry.RegisterCommand("objects:list", &ObjectsListCommand{}))
}
