//go:build std || multimap

package multimap

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func getMultiMap(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (*hazelcast.MultiMap, error) {
	name := ec.Props().GetString(base.FlagName)
	ci, err := cmd.ClientInternal(ctx, ec, sp)
	if err != nil {
		return nil, err
	}
	sp.SetText(fmt.Sprintf("Getting MultiMap '%s'", name))
	return ci.Client().GetMultiMap(ctx, name)
}

func makeDecodeResponseRowsFunc(decoder func(*hazelcast.ClientMessage) []*hazelcast.Data) func(context.Context, plug.ExecContext, *hazelcast.ClientMessage) ([]output.Row, error) {
	return func(ctx context.Context, ec plug.ExecContext, res *hazelcast.ClientMessage) ([]output.Row, error) {
		key := ec.GetStringArg(commands.ArgKey)
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		var rows []output.Row
		data := decoder(res)
		for _, r := range data {
			vt := r.Type()
			value, err := ci.DecodeData(*r)
			if err != nil {
				ec.Logger().Info("The value for %s was not decoded, due to error: %s", key, err.Error())
				value = serialization.NondecodedType(serialization.TypeToLabel(vt))
			}
			row := output.Row{
				output.Column{
					Name:  output.NameValue,
					Type:  vt,
					Value: value,
				},
			}
			if ec.Props().GetBool(base.FlagShowType) {
				row = append(row, output.Column{
					Name:  output.NameValueType,
					Type:  serialization.TypeString,
					Value: serialization.TypeToLabel(vt),
				})
			}
			rows = append(rows, row)
		}
		return rows, nil
	}
}
