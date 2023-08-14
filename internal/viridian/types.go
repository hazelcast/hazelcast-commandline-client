package viridian

type Cluster struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name"`
	HazelcastVersion   string       `json:"hazelcastVersion"`
	Memory             float64      `json:"memory"`
	MancenterURL       string       `json:"mancenterUrl"`
	MemoryUsage        float64      `json:"memoryUsage"`
	CreationTime       int64        `json:"creationTime"`
	StartTime          int64        `json:"startTime"`
	State              string       `json:"state"`
	DesiredState       string       `json:"desiredState"`
	ClientCount        int64        `json:"clientCount"`
	ClusterType        ClusterType  `json:"clusterType"`
	KubernetesClusters []K8sCluster `json:"kubernetesClusters"`
	HotBackupEnabled   bool         `json:"hotBackupEnabled"`
	HotRestartEnabled  bool         `json:"hotRestartEnabled"`
	PlanName           string       `json:"planName"`
	Regions            []Region     `json:"regions"`
	AllowedIps         []IP         `json:"allowedIps"`
	IPWhitelistEnabled bool         `json:"ipWhitelistEnabled"`
	MaxAvailableMemory int          `json:"maxAvailableMemory"`
}

type CustomClass struct {
	ID                       int64  `json:"id"`
	Name                     string `json:"name"`
	GeneratedFilename        string `json:"generatedFilename"`
	Status                   string `json:"status"`
	TemporaryCustomClassesId string `json:"temporaryCustomClassesId"`
}

type K8sCluster struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ClusterType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Region struct {
	Title string `json:"title"`
}

type IP struct {
	ID          int    `json:"id"`
	IP          string `json:"ip"`
	Description string `json:"description",omitempty`
}
