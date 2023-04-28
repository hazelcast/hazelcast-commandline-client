package prompt

import (
	"errors"
	"fmt"
	"io"
	"strings"

	gohxs "github.com/gohxs/readline"
)

type Prompter struct {
	Stdin           io.Reader
	Stdout          io.Writer
	InterruptPrompt string
}

func NewPrompter(r io.Reader, w io.Writer) *Prompter {
	return &Prompter{
		Stdin:           r,
		Stdout:          w,
		InterruptPrompt: "^C",
	}
}

func (p *Prompter) YesNoPrompt(q string) (bool, error) {
	prompt := fmt.Sprintf("%s (y/n) ", q)
	s, err := p.readline(prompt)
	if err != nil {
		return false, err
	}
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	switch s {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, errors.New("Invalid input")
	}
}

func (p *Prompter) readline(prompt string) (string, error) {
	cfg := &gohxs.Config{
		Prompt:                 prompt,
		InterruptPrompt:        p.InterruptPrompt,
		Stdout:                 p.Stdout,
		Stdin:                  p.Stdin,
		DisableAutoSaveHistory: true,
		HistoryLimit:           -1,
	}
	rl, err := gohxs.NewEx(cfg)
	if err != nil {
		return "", err
	}

	return rl.Readline()
}
