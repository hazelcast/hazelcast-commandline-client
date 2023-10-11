//go:build migration

package migration

import "github.com/hazelcast/hazelcast-go-client/types"

func MakeMigrationID() string {
	return types.NewUUID().String()
}
