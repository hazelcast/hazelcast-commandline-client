package prompt

import (
	"fmt"
	"io"
	"strings"

	gohxs "github.com/gohxs/readline"
)

type Prompt struct {
	stdin           io.Reader
	stdout          io.Writer
	interruptPrompt string
}

func New(r io.Reader, w io.Writer) *Prompt {
	return &Prompt{
		stdin:           r,
		stdout:          w,
		interruptPrompt: "^C",
	}
}

func (p *Prompt) YesNo(q string) (bool, error) {
	prompt := fmt.Sprintf("%s (y/n) ", q)
	s, err := p.readline(prompt, false)
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

func (p *Prompt) Text(q string) (string, error) {
	return p.readline(q, false)
}

func (p *Prompt) Password(q string) (string, error) {
	return p.readline(q, true)
}

func (p *Prompt) readline(prompt string, enableMask bool) (string, error) {
	// prevent closing the stdin
	stdin := io.NopCloser(p.stdin)
	cfg := &gohxs.Config{
		Prompt:                 prompt,
		InterruptPrompt:        p.interruptPrompt,
		Stdout:                 p.stdout,
		Stdin:                  stdin,
		DisableAutoSaveHistory: true,
		HistoryLimit:           -1,
		EnableMask:             enableMask,
		MaskRune:               '*',
	}
	rl, err := gohxs.NewEx(cfg)
	if err != nil {
		return "", fmt.Errorf("creating gohxs readline: %w", err)
	}
	defer rl.Close()
	return rl.Readline()
}
