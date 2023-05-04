package viridian

type Cluster struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	HazelcastVersion string `json:"hazelcastVersion"`
	State            string `json:"state"`
}

type CustomClass struct {
	Id                       int64  `json:"id"`
	Name                     string `json:"name"`
	GeneratedFilename        string `json:"generatedFilename"`
	Status                   string `json:"status"`
	TemporaryCustomClassesId string `json:"temporaryCustomClassesId"`
}
