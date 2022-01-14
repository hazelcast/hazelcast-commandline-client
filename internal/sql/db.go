package sql

import (
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"

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
	dsn := makeDSN(config)
	once.Do(func() {
		var err error
		if db, err = sql.Open("hazelcast", dsn); err != nil {
			panic(fmt.Sprintf("opening connection: %s", err.Error()))
		}
	})
	return db
}

func makeDSN(config *hazelcast.Config) string {
	ll := logger.OffLevel
	q := url.Values{}
	q.Add("cluster.name", config.Cluster.Name)
	q.Add("unisocket", strconv.FormatBool(config.Cluster.Unisocket))
	q.Add("log", string(ll))
	return fmt.Sprintf("hz://%s?%s", config.Cluster.Network.Addresses[0], q.Encode())
}
