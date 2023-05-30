package _set

import (
	"math"

	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type SetRemoveCommand struct{}

func (sc *SetRemoveCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(1, math.MaxInt)
	help := "Remove values to the given Set"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("set [value] [flags]")
	return nil
}
