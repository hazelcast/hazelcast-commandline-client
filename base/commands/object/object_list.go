//go:build std || object

package object

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/base/objects"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	Map                  = "map"
	ReplicatedMap        = "replicatedMap"
	MultiMap             = "multiMap"
	Queue                = "queue"
	Topic                = "topic"
	List                 = "list"
	Set                  = "set"
	PNCounter            = "pnCounter"
	FlakeIDGenerator     = "flakeIdGenerator"
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
	Cache,
	EventJournal,
	Ringbuffer,
	CardinalityEstimator,
}

const (
	flagShowHidden     = "show-hidden"
	argObjectType      = "objectType"
	argTitleObjectType = "object type"
)

type ListCommand struct{}

func (ListCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	long := fmt.Sprintf(`List distributed objects, optionally filter by type.
	
The object-type filter may be one of:
	
%s
CP objects such as AtomicLong cannot be listed.
`, objectFilterTypes())
	cc.SetCommandHelp(long, "List distributed objects")
	cc.AddBoolFlag(flagShowHidden, "", false, false, "show hidden and system objects")
	cc.AddStringSliceArg(argObjectType, argTitleObjectType, 0, 1)
	return nil
}

func (ListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	var typeFilter string
	fs := ec.GetStringSliceArg(argObjectType)
	if len(fs) > 0 {
		typeFilter = fs[0]
	}
	showHidden := ec.Props().GetBool(flagShowHidden)
	objs, err := objects.GetAll(ctx, ec, typeFilter, showHidden)
	if err != nil {
		return err
	}
	var rows []output.Row
	for _, o := range objs {
		valueCol := output.Column{
			Name:  "Object Name",
			Type:  serialization.TypeString,
			Value: o.Name,
		}
		if typeFilter != "" {
			rows = append(rows, output.Row{valueCol})
			continue
		}
		rows = append(rows, output.Row{
			output.Column{
				Name:  "Service Name",
				Type:  serialization.TypeString,
				Value: objects.ShortType(o.ServiceName),
			},
			valueCol,
		})
	}
	if len(rows) == 0 {
		ec.PrintlnUnnecessary("OK No objects found.")
		return nil
	}
	return ec.AddOutputRows(ctx, rows...)
}

func objectFilterTypes() string {
	var sb strings.Builder
	for _, o := range objTypes {
		sb.WriteString(fmt.Sprintf("\t* %s\n", o))
	}
	return sb.String()
}

func init() {
	// sort objectTypes so they look better in help
	sort.Slice(objTypes, func(i, j int) bool {
		return objTypes[i] < objTypes[j]
	})
	Must(plug.Registry.RegisterCommand("object:list", &ListCommand{}))
}
