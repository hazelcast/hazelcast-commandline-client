package migration

const (
	startQueueName     = "__datamigration_start_queue"
	statusMapName      = "__datamigration_1"
	statusMapEntryName = "status"
	statusComplete     = "COMPLETE"
	statusCanceled     = "CANCELED"
	statusFailed       = "FAILED"
)
