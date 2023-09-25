package terminal

import (
	"os"
	"strconv"

	"github.com/nathan-fiscaletti/consolesize-go"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
)

func IsPipe(v any) bool {
	s, ok := v.(Stater)
	if !ok {
		return false
	}
	fi, err := s.Stat()
	if err != nil {
		// do not activate pipe mode if there's a problem with getting stats of stdin
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

type Stater interface {
	Stat() (os.FileInfo, error)
}

func ConsoleWidth() int {
	if s, ok := os.LookupEnv(clc.EnvMaxCols); ok {
		v, err := strconv.Atoi(s)
		if err == nil {
			return v
		}
	}
	s, _ := consolesize.GetConsoleSize()
	if s == 0 {
		return 1000
	}
	return s
}
