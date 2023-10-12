//go:build std || migration

package migration

import (
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Initializer struct{}

func (Initializer) Init(cc plug.InitContext) error {
	cc.AddCommandGroup("migration", "Data Migration")
	return nil
}

func init() {
	plug.Registry.RegisterGlobalInitializer("01-migration", &Initializer{})
}
