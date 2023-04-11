package shell

import (
	"github.com/hazelcast/hazelcast-commandline-client/internal/terminal"
)

func IsPipe() bool {
	return terminal.IsPipe()
}
