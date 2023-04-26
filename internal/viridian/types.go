package viridian

type Cluster struct {
	ID               string
	Name             string
	HazelcastVersion string `json:"hazelcast_version"`
}
