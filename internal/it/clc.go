package it

func StartJetEnabledCluster(clusterName string) *TestCluster {
	port := NextPort()
	memberConfig := sqlXMLConfig(clusterName, "localhost", port)
	if SSLEnabled() {
		memberConfig = sqlXMLSSLConfig(clusterName, "localhost", port)
	}
	return StartNewClusterWithConfig(MemberCount(), memberConfig, port)
}
