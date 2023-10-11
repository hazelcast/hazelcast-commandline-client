//go:build migration

package migration

const (
	StatusMapName            = "__datamigration_migrations"
	MigrationsInProgressList = "__datamigrations_in_progress"
	CancelQueue              = "__datamigration_cancel_queue"
)

type Status string

const (
	StatusCanceling Status = "CANCELING"
)
