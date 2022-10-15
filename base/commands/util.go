package commands

import (
	"fmt"
	"io"
)

func printf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}
