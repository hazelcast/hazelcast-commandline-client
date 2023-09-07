package migration

const (
	StartQueueName           = "__datamigration_start_queue"
	StatusMapEntryName       = "status"
	UpdateTopic              = "__datamigration_updates_"
	MigrationsInProgressList = "__datamigrations_in_progress"
)

type Status string

const (
	StatusNone       Status = ""
	StatusComplete   Status = "COMPLETED"
	StatusCanceled   Status = "CANCELED"
	StatusFailed     Status = "FAILED"
	StatusInProgress Status = "IN_PROGRESS"
)
