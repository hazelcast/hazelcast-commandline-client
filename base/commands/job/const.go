package job

const (
	terminateModeRestartGraceful     int32 = 0
	terminateModeRestartForceful     int32 = 1
	terminateModeSuspendGraceful     int32 = 2
	terminateModeSuspendForceful     int32 = 3
	terminateModeCancelGraceful      int32 = 4
	terminateModeCancelForceful      int32 = 5
	flagForce                              = "force"
	flagIncludeSQL                         = "include-sql"
	flagIncludeUserCancelled               = "include-user-cancelled"
	flagName                               = "name"
	flagSnapshot                           = "snapshot"
	flagClass                              = "class"
	flagCancel                             = "cancel"
	statusNotRunning                       = 0
	statusStarting                         = 1
	statusRunning                          = 2
	statusSuspended                        = 3
	statusSuspendedExportingSnapshot       = 4
	statusCompleting                       = 5
	statusFailed                           = 6
	statusCompleted                        = 7
)
