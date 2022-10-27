package shell

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/lmorg/readline"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type EndLineFn func(line string) bool

type TextFn func(ctx context.Context, text string) error

type Shell struct {
	rl            *readline.Instance
	endLineFn     EndLineFn
	textFn        TextFn
	prompt1       string
	prompt2       string
	historyPath   string
	stderr        io.Writer
	stdout        io.Writer
	commentPrefix string
}

func New(prompt1, prompt2, historyPath string, stdout, stderr io.Writer, endLineFn EndLineFn, textFn TextFn) *Shell {
	rl := readline.NewInstance()
	if historyPath != "" {
		rl.History = NewFileHistory(historyPath)
	}
	rl.SetPrompt(prompt1)
	return &Shell{
		rl:            rl,
		endLineFn:     endLineFn,
		textFn:        textFn,
		prompt1:       prompt1,
		prompt2:       prompt2,
		historyPath:   historyPath,
		stderr:        stderr,
		stdout:        stdout,
		commentPrefix: "",
	}
}

func (sh *Shell) Close() error {
	if hs, ok := sh.rl.History.(*FileHistory); ok {
		hs.Close()
	}
	return nil
}

func (sh *Shell) SetCommentPrefix(pfx string) {
	sh.commentPrefix = pfx
}

func (sh *Shell) Start(ctx context.Context) error {
	for {
		text, err := sh.readTextReadline()
		if err == readline.CtrlC || err == readline.EOF {
			return nil
		}
		if err != nil {
			I2(fmt.Fprintf(sh.stderr, "%s\n", err.Error()))
		}
		if text == "" {
			continue
		}
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
		if err := sh.textFn(ctx, text); err != nil {
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				I2(fmt.Fprintf(sh.stderr, "%s\n", err.Error()))
			}
		}
		stop()
	}
}

func (sh *Shell) readTextReadline() (string, error) {
	prompt := sh.prompt1
	var sb strings.Builder
	for {
		sh.rl.SetPrompt(prompt)
		p, err := sh.rl.Readline()
		if err != nil {
			return "", err
		}
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if sh.commentPrefix != "" && strings.HasPrefix(p, sh.commentPrefix) {
			continue
		}
		sb.WriteString(p)
		sb.WriteString("\n")
		if sh.endLineFn(p) {
			break
		}
		prompt = sh.prompt2
	}
	text := sb.String()
	return text, nil
}

type FileHistory struct {
	lines      []string
	dirtyLines []string
	dirtyMu    *sync.Mutex
	path       string
	doneCh     chan struct{}
}

func NewFileHistory(path string) *FileHistory {
	h := &FileHistory{
		path:    path,
		dirtyMu: &sync.Mutex{},
		doneCh:  make(chan struct{}),
	}
	f, err := os.Open(path)
	if err == nil {
		// try to read the previous history items
		scn := bufio.NewScanner(f)
		for scn.Scan() {
			if scn.Err() != nil {
				break
			}
			h.lines = append(h.lines, strings.TrimSpace(scn.Text()))
		}
		f.Close()
	}
	go h.backgroundWriter()
	return h
}

func (hs *FileHistory) Close() {
	close(hs.doneCh)
}

func (hs *FileHistory) Write(s string) (int, error) {
	// add only unique lines
	if len(hs.lines) == 0 || s != hs.lines[len(hs.lines)-1] {
		// a unique line
		hs.lines = append(hs.lines, s)
		hs.dirtyMu.Lock()
		hs.dirtyLines = append(hs.dirtyLines, s)
		hs.dirtyMu.Unlock()
	}
	return len(hs.lines), nil
}

func (hs *FileHistory) GetLine(i int) (string, error) {
	if i >= len(hs.lines) {
		return "", fmt.Errorf("invalid history line: %d", i)
	}
	return hs.lines[i], nil
}

func (hs *FileHistory) Len() int {
	return len(hs.lines)
}

func (hs *FileHistory) Dump() interface{} {
	return hs.lines
}

func (hs *FileHistory) backgroundWriter() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-hs.doneCh:
			return
		case <-ticker.C:
			if err := hs.writeDirtyLines(); err != nil {
				return
			}
		}
	}
}

func (hs *FileHistory) writeDirtyLines() error {
	hs.dirtyMu.Lock()
	if len(hs.dirtyLines) == 0 {
		hs.dirtyMu.Unlock()
		return nil
	}
	cp := make([]string, len(hs.dirtyLines))
	copy(cp, hs.dirtyLines)
	hs.dirtyLines = nil
	hs.dirtyMu.Unlock()
	f, err := os.OpenFile(hs.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	bf := bufio.NewWriter(f)
	for _, line := range cp {
		_, err = bf.WriteString(line)
		if err != nil {
			return err
		}
		_, err = bf.WriteString("\n")
		if err != nil {
			return err
		}
	}
	// ignoring the error here
	_ = bf.Flush()
	return nil
}
