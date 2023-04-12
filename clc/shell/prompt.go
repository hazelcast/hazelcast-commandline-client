package shell

import (
	"fmt"
	"io"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

func Prompt(out io.Writer, in io.Reader, prompt string) (text string, err error) {
	check.I2(fmt.Fprint(out, prompt))
	if _, err = fmt.Fscanln(in, &text); err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func PasswordPrompt(out io.Writer, in io.Reader, prompt string) (text string, err error) {
	// XXX: the password can only be read from a real stdin
	check.I2(fmt.Fprint(out, prompt))
	b, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
