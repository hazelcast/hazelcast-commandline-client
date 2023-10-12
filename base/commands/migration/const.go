//go:build migration

package migration

const (
	StartQueueName           = "__datamigration_start_queue"
	StatusMapEntryName       = "status"
	StatusMapName            = "__datamigration_migrations"
	UpdateTopicPrefix        = "__datamigration_updates_"
	DebugLogsListPrefix      = "__datamigration_debug_logs_"
	MigrationsInProgressList = "__datamigrations_in_progress"
	startQueueName           = "__datamigration_start_queue"
	statusMapEntryName       = "status"
	argDMTConfig             = "dmtConfig"
	argTitleDMTConfig        = "DMT configuration"
	CancelQueue              = "__datamigration_cancel_queue"
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
