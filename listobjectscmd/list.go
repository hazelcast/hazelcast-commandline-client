package listobjectscmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/connection"
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

var dsObjTypes = []string{
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

func New(config *hazelcast.Config) *cobra.Command {
	var (
		objectType string
		showHidden bool
	)
	cmd := cobra.Command{
		Use:   "list-objects [distributed object type]",
		Short: "List distributed objects in the cluster",
		Long:  `List type and name of the distributed objects present in the cluster`,
		Example: `  list
  list --type map
  list --type fencedlock`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ConnectToCluster(ctx, config)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Can not connect to the cluster")
			}
			list, err := getObjects(ctx, c, objectType, showHidden)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Can not get the distributed objects information")
			}
			sort.Strings(list)
			l := strings.Join(list, "\n")
			cmd.Println(l)
			return nil
		},
	}
	cmd.Flags().StringVarP(&objectType, "type", "t", "", fmt.Sprintf("type: %s", strings.Join(dsObjTypes, ",")))
	cmd.Flags().BoolVar(&showHidden, "show-hidden", false, `show objects with names starting with "__"`)
	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return dsObjTypes, cobra.ShellCompDirectiveDefault
	})
	return &cmd
}

func getObjects(ctx context.Context, c *hazelcast.Client, filter string, showHidden bool) ([]string, error) {
	ts, err := c.GetDistributedObjectsInfo(ctx)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, t := range ts {
		if !showHidden && strings.HasPrefix(t.Name, "__") {
			continue
		}
		toFilter := strings.TrimPrefix(t.ServiceName, "hz:impl:")
		toFilter = strings.TrimSuffix(toFilter, "Service")
		if filter == "" {
			names = append(names, fmt.Sprintf("%s %s", toFilter, t.Name))
			continue
		}
		if filter == toFilter {
			names = append(names, t.Name)
		}
	}
	return names, nil
}
