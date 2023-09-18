package migration

const (
	StartQueueName           = "__datamigration_start_queue"
	StatusMapEntryName       = "status"
	StatusMapPrefix          = "__datamigration_"
	UpdateTopicPrefix        = "__datamigration_updates_"
	DebugLogsListPrefix      = "__datamigration_debug_logs_"
	MigrationsInProgressList = "__datamigrations_in_progress"
)

type Status string

const (
	StatusStarted    Status = "STARTED"
	Canceling        Status = "CANCELING"
	StatusComplete   Status = "COMPLETED"
	StatusCanceled   Status = "CANCELED"
	StatusFailed     Status = "FAILED"
	StatusInProgress Status = "IN_PROGRESS"
)

const flagOutputDir = "output-dir"
