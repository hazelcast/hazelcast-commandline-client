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
	long := fmt.Sprintf(`List distributed objects, optionally filter by type.
	
The object-type filter may be one of:
%s
`, objectFilterTypes())
	cc.SetCommandHelp(long, "List distributed objects")
	cc.AddBoolFlag(flagShowHidden, "", false, false, "show hidden and system objects")
	cc.SetPositionalArgCount(0, 1)
	return nil
}

func (cm ObjectsListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	var typeFilter string
	if len(ec.Args()) > 0 {
		typeFilter = ec.Args()[0]
	}
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

func objectFilterTypes() string {
	var sb strings.Builder
	for _, o := range objTypes {
		sb.WriteString(fmt.Sprintf("\t* %s\n", strings.ToLower(o)))
	}
	return sb.String()
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
	typeFilter = strings.ToLower(typeFilter)
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
	s := strings.TrimSuffix(strings.TrimPrefix(svcName, "hz:impl:"), "Service")
	return strings.ToLower(s)
}

func init() {
	// sort objectTypes so they look better in help
	sort.Slice(objTypes, func(i, j int) bool {
		return objTypes[i] < objTypes[j]
	})
	Must(plug.Registry.RegisterCommand("objects:list", &ObjectsListCommand{}))
}
