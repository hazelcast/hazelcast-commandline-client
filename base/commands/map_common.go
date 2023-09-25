package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type mapPrefixFunc[T any] func(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (T, error)

type nameRequestEncodeFunc func(name string) *hazelcast.ClientMessage
type pairsResponseDecodeFunc func(message *hazelcast.ClientMessage) []hazelcast.Pair

type MapEntrySetCommand[T any] struct {
	typeName string
	encoder  nameRequestEncodeFunc
	decoder  pairsResponseDecodeFunc
	prefixer mapPrefixFunc[T]
}

func NewMapEntrySetCommand[T any](typeName string, encoder nameRequestEncodeFunc, decoder pairsResponseDecodeFunc, prefixer mapPrefixFunc[T]) *MapEntrySetCommand[T] {
	return &MapEntrySetCommand[T]{
		typeName: typeName,
		encoder:  encoder,
		decoder:  decoder,
		prefixer: prefixer,
	}
}

func (cm MapEntrySetCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("entry-set")
	help := fmt.Sprintf("Get all entries of a %s", cm.typeName)
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm MapEntrySetCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		if _, err = cm.prefixer(ctx, ec, sp); err != nil {
			return nil, err
		}
		req := cm.encoder(name)
		sp.SetText(fmt.Sprintf("Getting entries of %s", name))
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		pairs := cm.decoder(resp)
		rows := output.DecodePairs(ci, pairs, showType)
		return rows, nil
	})
	if err != nil {
		return err
	}
	stop()
	return AddDDSRows(ctx, ec, cm.typeName, "entries", rowsV.([]output.Row))
}

type getRequestEncodeFunc func(name string, keyData hazelcast.Data, threadID int64) *hazelcast.ClientMessage
type getResponseDecodeFunc func(ctx context.Context, ec plug.ExecContext, res *hazelcast.ClientMessage) ([]output.Row, error)

type MapGetCommand[T any] struct {
	typeName string
	encoder  getRequestEncodeFunc
	decoder  getResponseDecodeFunc
	prefixer mapPrefixFunc[T]
}

func NewMapGetCommand[T any](typeName string, encoder getRequestEncodeFunc, decoder getResponseDecodeFunc, prefixer mapPrefixFunc[T]) *MapGetCommand[T] {
	return &MapGetCommand[T]{
		typeName: typeName,
		encoder:  encoder,
		decoder:  decoder,
		prefixer: prefixer,
	}
}

func (cm MapGetCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("get")
	AddKeyTypeFlag(cc)
	help := fmt.Sprintf("Get a value from the given %s", cm.typeName)
	cc.SetCommandHelp(help, help)
	cc.AddStringArg(ArgKey, ArgTitleKey)
	return nil
}

func (cm MapGetCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	keyStr := ec.GetStringArg(ArgKey)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		if _, err = cm.prefixer(ctx, ec, sp); err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Getting from %s '%s'", cm.typeName, name))
		keyData, err := MakeKeyData(ec, ci, keyStr)
		if err != nil {
			return nil, err
		}
		req := cm.encoder(name, keyData, 0)
		resp, err := ci.InvokeOnKey(ctx, req, keyData, nil)
		if err != nil {
			return nil, err
		}
		return cm.decoder(ctx, ec, resp)
	})
	if err != nil {
		return err
	}
	stop()
	rows := rowsV.([]output.Row)
	if len(rows) == 0 {
		ec.PrintlnUnnecessary("OK No values.")
		return nil
	}
	return ec.AddOutputRows(ctx, rowsV.([]output.Row)...)
}

type dataSliceDecoderFunc func(message *hazelcast.ClientMessage) []*hazelcast.Data

type MapKeySetCommand[T any] struct {
	typeName string
	encoder  nameRequestEncodeFunc
	decoder  dataSliceDecoderFunc
	prefixer mapPrefixFunc[T]
}

func NewMapKeySetCommand[T any](typeName string, encoder nameRequestEncodeFunc, decoder dataSliceDecoderFunc, prefixer mapPrefixFunc[T]) *MapKeySetCommand[T] {
	return &MapKeySetCommand[T]{
		typeName: typeName,
		encoder:  encoder,
		decoder:  decoder,
		prefixer: prefixer,
	}
}

func (cm MapKeySetCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("key-set")
	help := fmt.Sprintf("Get all keys of a %s", cm.typeName)
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm MapKeySetCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		if _, err = cm.prefixer(ctx, ec, sp); err != nil {
			return nil, err
		}
		req := cm.encoder(name)
		sp.SetText(fmt.Sprintf("Getting keys of %s '%s'", cm.typeName, name))
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		data := cm.decoder(resp)
		var rows []output.Row
		for _, r := range data {
			var row output.Row
			t := r.Type()
			v, err := ci.DecodeData(*r)
			if err != nil {
				v = serialization.NondecodedType(serialization.TypeToLabel(t))
			}
			row = append(row, output.NewKeyColumn(t, v))
			if showType {
				row = append(row, output.NewKeyTypeColumn(t))
			}
			rows = append(rows, row)
		}
		return rows, nil
	})
	if err != nil {
		return err
	}
	stop()
	return AddDDSRows(ctx, ec, cm.typeName, "keys", rowsV.([]output.Row))
}

type MapRemoveCommand struct {
	typeName string
	encoder  getRequestEncodeFunc
	decoder  getResponseDecodeFunc
}

func NewMapRemoveCommand(typeName string, encoder getRequestEncodeFunc, decoder getResponseDecodeFunc) *MapRemoveCommand {
	return &MapRemoveCommand{
		typeName: typeName,
		encoder:  encoder,
		decoder:  decoder,
	}
}

func (cm MapRemoveCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("remove")
	AddKeyTypeFlag(cc)
	help := fmt.Sprintf("Remove a value from the given %s", cm.typeName)
	cc.SetCommandHelp(help, help)
	cc.AddStringArg(ArgKey, ArgTitleKey)
	return nil
}

func (cm MapRemoveCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	keyStr := ec.GetStringArg(ArgKey)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Removing from %s '%s'", cm.typeName, name))
		keyData, err := MakeKeyData(ec, ci, keyStr)
		if err != nil {
			return nil, err
		}
		req := cm.encoder(name, keyData, 0)
		resp, err := ci.InvokeOnKey(ctx, req, keyData, nil)
		if err != nil {
			return nil, err
		}
		return cm.decoder(ctx, ec, resp)
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Removed the entry from %s '%s'.\n", cm.typeName, name)
	ec.PrintlnUnnecessary(msg)
	return ec.AddOutputRows(ctx, rowsV.([]output.Row)...)
}

type Locker interface {
	LockWithLease(ctx context.Context, key any, leaseTime time.Duration) error
	Lock(ctx context.Context, key any) error
}

type getLockerFunc[T Locker] func(context.Context, plug.ExecContext, clc.Spinner) (T, error)

type LockCommand[T Locker] struct {
	typeName string
	getFn    getLockerFunc[T]
}

func NewLockCommand[T Locker](typeName string, getFn getLockerFunc[T]) *LockCommand[T] {
	return &LockCommand[T]{
		typeName: typeName,
		getFn:    getFn,
	}
}

func (cm LockCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("lock")
	long := fmt.Sprintf(`Lock a key in the given %s

This command is only available in the interactive mode.`, cm.typeName)
	short := fmt.Sprintf("Lock a key in the given %s", cm.typeName)
	cc.SetCommandHelp(long, short)
	AddKeyTypeFlag(cc)
	cc.AddIntFlag(FlagTTL, "", clc.TTLUnset, false, "time-to-live (ms)")
	cc.AddStringArg(ArgKey, ArgTitleKey)
	return nil
}

func (cm LockCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		key, err := MakeValueFromString(ec.GetStringArg(ArgKey), ec.Props().GetString(FlagKeyType))
		if err != nil {
			return nil, err
		}
		m, err := cm.getFn(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Locking the key in %s '%s'", cm.typeName, name))
		if ttl := GetTTL(ec); ttl != clc.TTLUnset {
			return nil, m.LockWithLease(ctx, key, time.Duration(GetTTL(ec)))
		}
		return nil, m.Lock(ctx, key)
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Locked the key in %s '%s'.", cm.typeName, name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

type LockTrier interface {
	TryLockWithLease(ctx context.Context, key any, leaseTime time.Duration) (bool, error)
	TryLock(ctx context.Context, key any) (bool, error)
}

type getLockTrierFunc[T LockTrier] func(context.Context, plug.ExecContext, clc.Spinner) (T, error)

type MapTryLockCommand[T LockTrier] struct {
	typeName string
	getFn    getLockTrierFunc[T]
}

func NewTryLockCommand[T LockTrier](typeName string, getFn getLockTrierFunc[T]) *MapTryLockCommand[T] {
	return &MapTryLockCommand[T]{
		typeName: typeName,
		getFn:    getFn,
	}
}

func (cm MapTryLockCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("try-lock")
	long := fmt.Sprintf(`Try to lock a key in the given %s

Returns the result without waiting for the lock to be unlocked.

This command is only available in the interactive mode.`, cm.typeName)
	short := fmt.Sprintf("Try to lock a key in the given %s", cm.typeName)
	cc.SetCommandHelp(long, short)
	AddKeyTypeFlag(cc)
	cc.AddIntFlag(FlagTTL, "", clc.TTLUnset, false, "time-to-live (ms)")
	cc.AddStringArg(ArgKey, ArgTitleKey)
	return nil
}

func (cm MapTryLockCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		mapName := ec.Props().GetString(base.FlagName)
		sp.SetText(fmt.Sprintf("Locking key in map %s", mapName))
		m, err := cm.getFn(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		key, err := MakeValueFromString(ec.GetStringArg(ArgKey), ec.GetStringArg(FlagKeyType))
		if err != nil {
			return nil, err
		}
		var locked bool
		if ttl := GetTTL(ec); ttl != clc.TTLUnset {
			locked, err = m.TryLockWithLease(ctx, key, time.Duration(GetTTL(ec)))
		} else {
			locked, err = m.TryLock(ctx, key)
		}
		row := output.Row{
			{
				Name:  "Locked",
				Type:  serialization.TypeBool,
				Value: locked,
			},
		}
		if ec.Props().GetBool(base.FlagShowType) {
			row = append(row, output.Column{
				Name:  "Type",
				Type:  serialization.TypeString,
				Value: serialization.TypeToLabel(serialization.TypeBool),
			})
		}
		return []output.Row{row}, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, rowsV.([]output.Row)...)
}

type Unlocker interface {
	Unlock(ctx context.Context, key any) error
}

type getUnlockerFunc[T Unlocker] func(context.Context, plug.ExecContext, clc.Spinner) (T, error)

type MapUnlockCommand[T Unlocker] struct {
	typeName string
	getFn    getUnlockerFunc[T]
}

func NewMapUnlockCommand[T Unlocker](typeName string, getFn getUnlockerFunc[T]) *MapUnlockCommand[T] {
	return &MapUnlockCommand[T]{
		typeName: typeName,
		getFn:    getFn,
	}
}

func (cm MapUnlockCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("unlock")
	long := fmt.Sprintf(`Unlock a key in the given %s

This command is only available in the interactive mode.`, cm.typeName)
	short := fmt.Sprintf("Unlock a key in the given %s", cm.typeName)
	cc.SetCommandHelp(long, short)
	AddKeyTypeFlag(cc)
	cc.AddStringArg(ArgKey, ArgTitleKey)
	return nil
}

func (cm MapUnlockCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		key, err := MakeValueFromString(ec.GetStringArg(ArgKey), ec.Props().GetString(FlagKeyType))
		sp.SetText(fmt.Sprintf("Unlocking key in %s '%s'", cm.typeName, name))
		m, err := cm.getFn(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		return nil, m.Unlock(ctx, key)
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Unlocked the key in %s '%s'.", cm.typeName, name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

type MapValuesCommand[T any] struct {
	typeName string
	encoder  nameRequestEncodeFunc
	decoder  dataSliceDecoderFunc
	prefixer mapPrefixFunc[T]
}

func NewMapValuesCommand[T any](typeName string, encoder nameRequestEncodeFunc, decoder dataSliceDecoderFunc, prefixer mapPrefixFunc[T]) *MapValuesCommand[T] {
	return &MapValuesCommand[T]{
		typeName: typeName,
		encoder:  encoder,
		decoder:  decoder,
		prefixer: prefixer,
	}
}

func (cm MapValuesCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("values")
	help := fmt.Sprintf("Get all values of a %s", cm.typeName)
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm *MapValuesCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	showType := ec.Props().GetBool(base.FlagShowType)
	rowsV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		if _, err := cm.prefixer(ctx, ec, sp); err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Getting values of %s", name))
		req := cm.encoder(name)
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		data := cm.decoder(resp)
		var rows []output.Row
		for _, r := range data {
			var row output.Row
			t := r.Type()
			v, err := ci.DecodeData(*r)
			if err != nil {
				v = serialization.NondecodedType(serialization.TypeToLabel(t))
			}
			row = append(row, output.NewValueColumn(t, v))
			if showType {
				row = append(row, output.NewValueTypeColumn(t))
			}
			rows = append(rows, row)
		}
		return rows, nil
	})
	if err != nil {
		return err
	}
	stop()
	return AddDDSRows(ctx, ec, cm.typeName, "values", rowsV.([]output.Row))
}
