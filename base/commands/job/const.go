//go:build std || job

package job

const (
	flagForce                = "force"
	flagIncludeSQL           = "include-sql"
	flagIncludeUserCancelled = "include-user-cancelled"
	flagName                 = "name"
	flagSnapshot             = "snapshot"
	flagClass                = "class"
	flagCancel               = "cancel"
	flagRetries              = "retries"
	flagWait                 = "wait"
	argJobID                 = "jobID"
	argTitleJobID            = "job ID or name"
)
