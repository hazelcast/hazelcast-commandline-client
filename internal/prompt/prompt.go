package prompt

import (
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
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
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
		return "", fmt.Errorf("creating gohxs readline: %w", err)
	}

	return rl.Readline()
}
