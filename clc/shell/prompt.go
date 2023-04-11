package shell

import (
	"fmt"
	"io"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func Prompt(out io.Writer, in io.Reader, prompt string) (text string, err error) {
	check.I2(fmt.Fprint(out, prompt))
	if _, err = fmt.Fscanln(in, &text); err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
