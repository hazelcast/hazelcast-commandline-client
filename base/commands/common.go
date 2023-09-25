package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	_ "github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	FlagKeyType   = "key-type"
	FlagValueType = "value-type"
	ArgKey        = "key"
	ArgTitleKey   = "key"
)

type Destroyer interface {
	Destroy(ctx context.Context) error
}

type getDestroyerFunc[T Destroyer] func(context.Context, plug.ExecContext, clc.Spinner) (T, error)

type DestroyCommand[T Destroyer] struct {
	typeName       string
	getDestroyerFn getDestroyerFunc[T]
}

func NewDestroyCommand[T Destroyer](typeName string, getFn getDestroyerFunc[T]) *DestroyCommand[T] {
	return &DestroyCommand[T]{
		typeName:       typeName,
		getDestroyerFn: getFn,
	}
}

func (cm DestroyCommand[T]) Init(cc plug.InitContext) error {
	long := fmt.Sprintf(`Destroy a %s

This command will delete the %s and the data in it will not be available anymore.`, cm.typeName, cm.typeName)
	cc.SetCommandUsage("destroy")
	short := fmt.Sprintf("Destroy a %s", cm.typeName)
	cc.SetCommandHelp(long, short)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the destroy operation")
	return nil
}

func (cm DestroyCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo(fmt.Sprintf("%s '%s' will be deleted irreversibly, proceed?", cm.typeName, name))
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		m, err := cm.getDestroyerFn(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Destroying %s '%s'", cm.typeName, name))
		if err := m.Destroy(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Destroyed %s '%s'.", cm.typeName, name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

type Clearer interface {
	Clear(ctx context.Context) error
}

type getClearerFunc[T Clearer] func(context.Context, plug.ExecContext, clc.Spinner) (T, error)

type ClearCommand[T Clearer] struct {
	typeName     string
	getClearerFn getClearerFunc[T]
}

func NewClearCommand[T Clearer](typeName string, getFn getClearerFunc[T]) *ClearCommand[T] {
	return &ClearCommand[T]{
		typeName:     typeName,
		getClearerFn: getFn,
	}
}

func (cm ClearCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("clear")
	help := fmt.Sprintf("Deletes all entries of a %s", cm.typeName)
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(clc.FlagAutoYes, "", false, false, "skip confirming the clear operation")
	return nil
}

func (cm ClearCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	autoYes := ec.Props().GetBool(clc.FlagAutoYes)
	if !autoYes {
		p := prompt.New(ec.Stdin(), ec.Stdout())
		yes, err := p.YesNo(fmt.Sprintf("Content of %s '%s' will be deleted irreversibly. Proceed?", cm.typeName, name))
		if err != nil {
			ec.Logger().Info("User input could not be processed due to error: %s", err.Error())
			return errors.ErrUserCancelled
		}
		if !yes {
			return errors.ErrUserCancelled
		}
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		m, err := cm.getClearerFn(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Clearing %s '%s'", cm.typeName, name))
		if err := m.Clear(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Cleared %s '%s'.", cm.typeName, name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

type Sizer interface {
	Size(ctx context.Context) (int, error)
}

type getSizerFunc[T Sizer] func(context.Context, plug.ExecContext, clc.Spinner) (T, error)

type SizeCommand[T Sizer] struct {
	typeName   string
	getSizerFn getSizerFunc[T]
}

func NewSizeCommand[T Sizer](typeName string, getFn getSizerFunc[T]) *SizeCommand[T] {
	return &SizeCommand[T]{
		typeName:   typeName,
		getSizerFn: getFn,
	}
}

func (cm SizeCommand[T]) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("size")
	help := fmt.Sprintf("Returns the size of the given %s", cm.typeName)
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm SizeCommand[T]) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		m, err := cm.getSizerFn(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Getting the size of %s '%s'", cm.typeName, name))
		return m.Size(ctx)
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, output.Row{
		{
			Name:  "Size",
			Type:  serialization.TypeInt32,
			Value: int32(sv.(int)),
		},
	})
}

func AddKeyTypeFlag(cc plug.InitContext) {
	help := fmt.Sprintf("key type (one of: %s)", strings.Join(internal.SupportedTypeNames, ", "))
	cc.AddStringFlag(base.FlagKeyType, "k", "string", false, help)
}

func AddValueTypeFlag(cc plug.InitContext) {
	help := fmt.Sprintf("value type (one of: %s)", strings.Join(internal.SupportedTypeNames, ", "))
	cc.AddStringFlag(FlagValueType, "v", "string", false, help)
}

func MakeKeyData(ec plug.ExecContext, ci *hazelcast.ClientInternal, keyStr string) (hazelcast.Data, error) {
	kt := ec.Props().GetString(FlagKeyType)
	if kt == "" {
		kt = "string"
	}
	key, err := mk.ValueFromString(keyStr, kt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(key)
}

func MakeValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, valueStr string) (hazelcast.Data, error) {
	vt := ec.Props().GetString(base.FlagValueType)
	if vt == "" {
		vt = "string"
	}
	value, err := mk.ValueFromString(valueStr, vt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(value)
}

func MakeKeyValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, keyStr, valueStr string) (hazelcast.Data, hazelcast.Data, error) {
	kd, err := MakeKeyData(ec, ci, keyStr)
	if err != nil {
		return nil, nil, err
	}
	vd, err := MakeValueData(ec, ci, valueStr)
	if err != nil {
		return nil, nil, err
	}
	return kd, vd, nil
}

func AddDDSRows(ctx context.Context, ec plug.ExecContext, typeName, what string, rows []output.Row) error {
	name := ec.Props().GetString(base.FlagName)
	if len(rows) == 0 {
		msg := fmt.Sprintf("OK No %s found in %s '%s'.", what, typeName, name)
		ec.PrintlnUnnecessary(msg)
		return nil

	}
	return ec.AddOutputRows(ctx, rows...)
}

func GetTTL(ec plug.ExecContext) int64 {
	if _, ok := ec.Props().Get(FlagTTL); ok {
		return ec.Props().GetInt(FlagTTL)
	}
	return clc.TTLUnset
}

func IsYes(ec plug.ExecContext) bool {
	yes := ec.Props().GetBool(clc.FlagAutoYes)
	if yes {
		return true
	}
	return os.Getenv(clc.EnvYes) == "1"
}
