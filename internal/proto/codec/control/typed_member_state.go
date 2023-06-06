package control

type TimedMemberStateWrapper struct {
	TimedMemberState TimedMemberState `json:"timedMemberState"`
}

type TimedMemberState struct {
	MemberState MemberState `json:"memberState"`
	Master      bool        `json:"master"`
	ClusterName string      `json:"clusterName"`
}

type MemberState struct {
	Address   string    `json:"address"`
	Uuid      string    `json:"uuid"`
	Name      string    `json:"name"`
	NodeState NodeState `json:"nodeState"`
}

type NodeState struct {
	State         string `json:"nodeState"`
	MemberVersion string `json:"memberVersion"`
}
