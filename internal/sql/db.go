package sql

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/hazelcast/hazelcast-go-client/sql/driver"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var (
	db   *sql.DB
	once sync.Once
)

func Get() *sql.DB {
	config, err := internal.MakeConfig()
	// TODO error look like unhandled although it is handled in MakeConfig. Find a better approach
	if err != nil {
		panic(fmt.Sprintf("cannot process config %s", err.Error()))
	}
	once.Do(func() {
		var err error
		if db = driver.Open(*config); err != nil {
			panic(fmt.Sprintf("opening connection: %s", err.Error()))
		}
	})
	return db
}
