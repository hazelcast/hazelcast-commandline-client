package internal

const (
	TypeString = "string"
	TypeJSON   = "json"
)

const (
	ClusterGetStateEndpoint    = "/hazelcast/rest/management/cluster/state"
	ClusterChangeStateEndpoint = "/hazelcast/rest/management/cluster/changeState"
	ClusterShutdownEndpoint    = "/hazelcast/rest/management/cluster/clusterShutdown"
	ClusterVersionEndpoint     = "/hazelcast/rest/management/cluster/version"
)

const (
	ClusterGetState    = "get-state"
	ClusterChangeState = "change-state"
	ClusterShutdown    = "shutdown"
	ClusterVersion     = "version"
)

const (
	ClusterStateActive      = "active"
	ClusterStateNoMigration = "no_migration"
	ClusterStatePassive     = "passive"
	ClusterStateFrozen      = "frozen"
)
