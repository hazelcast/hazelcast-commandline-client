package terminal

import "os"

func IsPipe() bool {
	// TODO: parameterize os.Stdin
	fi, err := os.Stdin.Stat()
	if err != nil {
		// do not activate pipe mode if there's a problem with getting stats of stdin
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}
