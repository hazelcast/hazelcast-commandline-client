package clc

const (
	ShortcutConfig                = "c"
	ShortcutFormat                = "f"
	PropertyClusterAddress        = "cluster.address"
	PropertyClusterName           = "cluster.name"
	PropertyClusterDiscoveryToken = "cluster.discovery-token"
	PropertyClusterUser           = "cluster.user"
	PropertyClusterPassword       = "cluster.password"
	PropertyClusterAPIBase        = "cluster.api-base"
	PropertyFormat                = "format"
	PropertyVerbose               = "verbose"
	PropertyQuiet                 = "quiet"
	PropertyTimeout               = "timeout"
	// PropertyConfig is the config name or path
	// TODO: Separate config name and path
	PropertyConfig              = "config"
	PropertyLogLevel            = "log.level"
	PropertyLogPath             = "log.path"
	PropertySSLEnabled          = "ssl.enabled"
	PropertySSLServerName       = "ssl.server"
	PropertySSLCAPath           = "ssl.ca-path"
	PropertySSLCertPath         = "ssl.cert-path"
	PropertySSLKeyPath          = "ssl.key-path"
	PropertySSLKeyPassword      = "ssl.key-password"
	PropertySSLSkipVerify       = "ssl.skip-verify"
	PropertyExperimentalAPIBase = "experimental.api-base"
	GroupDDSID                  = "dds"
	GroupDDSTitle               = "Distributed Data Structures"
	GroupJetID                  = "jet"
	GroupJetTitle               = "Jet"
	EnvMaxCols                  = "CLC_MAX_COLS"
	EnvConfig                   = "CLC_CONFIG"
	EnvSkipServerVersionCheck   = "CLC_SKIP_SERVER_VERSION_CHECK"
	FlagAutoYes                 = "yes"
	MaxArgs                     = 65535
)
