//go:build std || migration

package migration

const (
	StartQueueName      = "__datamigration_start_queue"
	EstimateQueueName   = "__datamigration_estimate_queue"
	StatusMapName       = "__datamigration_migrations"
	DebugLogsListPrefix = "__datamigration_debug_logs_"
	CancelQueue         = "__datamigration_cancel_queue"
	argDMTConfig        = "dmtConfig"
	argTitleDMTConfig   = "DMT configuration"
)

type Status string

const (
	StatusStarted    Status = "STARTED"
	StatusCanceling  Status = "CANCELING"
	StatusComplete   Status = "COMPLETED"
	StatusCanceled   Status = "CANCELED"
	StatusFailed     Status = "FAILED"
	StatusInProgress Status = "IN_PROGRESS"
)

const flagOutputDir = "output-dir"

const banner = `Hazelcast Data Migration Tool v5.3.0
(c) 2023 Hazelcast, Inc.
`
