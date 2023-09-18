package metric

type Query struct {
	Date                       string `json:"date"`
	ID                         string `json:"id"`
	Version                    string `json:"version"`
	AcquisitionSource          string `json:"acquisitionSource"`
	Architecture               string `json:"architecture"`
	OS                         string `json:"os"`
	ClusterUUID                string `json:"clusterUUID"`
	ViridianClusterID          string `json:"viridianClusterID"`
	ClusterConfigCount         int    `json:"clusterConfigCount"`
	SqlRunCount                int    `json:"sqlRunCount"`
	MapRunCount                int    `json:"mapRunCount"`
	TopicRunCount              int    `json:"topicRunCount"`
	QueueRunCount              int    `json:"queueRunCount"`
	MultiMapRunCont            int    `json:"multiMapRunCount"`
	ListRunCount               int    `json:"listRunCount"`
	DemoRunCount               int    `json:"demoRunCount"`
	ProjectRunCount            int    `json:"projectRunCount"`
	JobRunCount                int    `json:"jobRunCount"`
	ViridianRunCount           int    `json:"viridianRunCount"`
	SetRunCount                int    `json:"setRunCount"`
	ShellRunCount              int    `json:"shellRunCount"`
	ScriptRunCount             int    `json:"scriptRunCount"`
	TotalRunCount              int    `json:"totalRunCount"`
	AtomicLongRunCount         int    `json:"atomicLongRunCount"`
	InteractiveModeRunCount    int    `json:"interactiveModeRunCount"`
	NoninteractiveModeRunCount int    `json:"noninteractiveModeRunCount"`
	ScriptModeRunCount         int    `json:"scriptModeRunCount"`
}
