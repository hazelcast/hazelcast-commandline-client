package terminal

import "os"

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
