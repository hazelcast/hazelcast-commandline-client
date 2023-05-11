package viridian

type Cluster struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	HazelcastVersion string `json:"hazelcastVersion"`
	State            string `json:"state"`
}

type CustomClass struct {
	ID                       int64  `json:"id"`
	Name                     string `json:"name"`
	GeneratedFilename        string `json:"generatedFilename"`
	Status                   string `json:"status"`
	TemporaryCustomClassesId string `json:"temporaryCustomClassesId"`
}

type K8sCluster struct {
	ID int `json:"id"`
}

type ClusterType int

const (
	ClusterTypeDev  ClusterType = 6
	ClusterTypeProd ClusterType = 5
)

type ClusterPlan string

const (
	ClusterPlanServerless ClusterPlan = "SERVERLESS"
)
