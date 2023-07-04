package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands/object"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
)

func Indexes(ctx context.Context, ec plug.ExecContext, mapName string) error {
	/*
		err := addIndex(ctx, ec)
		if err != nil {
			return err
		}
	*/
	var mapNames []string
	if mapName != "" {
		mapNames = append(mapNames, mapName)
	} else {
		maps, err := object.GetObjects(ctx, ec, object.Map, false)
		if err != nil {
			return err
		}
		for _, mm := range maps {
			mapNames = append(mapNames, mm.Name)
		}
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	resp, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		allIndexes := make(map[string][]types.IndexConfig)
		for _, mapName := range mapNames {
			sp.SetText(fmt.Sprintf("Getting indexes of map %s", mapName))
			req := codec.EncodeMCGetMapConfigRequest(mapName)
			resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
			if err != nil {
				return nil, err
			}
			_, _, _, _, _, _, _, _, _, _, globalIndexes := codec.DecodeMCGetMapConfigResponse(resp)
			if err != nil {
				return nil, err
			}
			allIndexes[mapName] = globalIndexes
		}
		return allIndexes, nil
	})
	stop()
	var rows []output.Row
	for mapName, indexes := range resp.(map[string][]types.IndexConfig) {
		for _, index := range indexes {
			rows = append(rows, output.Row{output.Column{
				Name:  "Map Name",
				Type:  serialization.TypeString,
				Value: mapName,
			}, output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: index.Name,
			}, output.Column{
				Name:  "Attributes",
				Type:  serialization.TypeStringArray,
				Value: index.Attributes,
			}})
		}
	}
	return ec.AddOutputRows(ctx, rows...)
}
