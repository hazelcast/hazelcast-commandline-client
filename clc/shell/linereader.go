package shell

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/fatih/color"
	gohxs "github.com/gohxs/readline"
	ny "github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/simplehistory"
)

type LineReader interface {
	SetPrompt(prompt string)
	ReadLine(ctx context.Context) (string, error)
	Close() error
	AddToHistory(item string)
}

func (sh *Shell) createNyLineReader(prompt string) error {
	history := simplehistory.New()
	ed := ny.Editor{
		Prompt: func() (int, error) {
			return fmt.Fprint(sh.stdout, prompt)
		},
		Writer:         sh.stdout,
		History:        history,
		HistoryCycling: true,
		Coloring:       &SQLColoring{},
	}
	sh.lr = NewNyLineReader(&ed)
	return nil
}

type NyLineReader struct {
	ed   *ny.Editor
	hist *simplehistory.Container
}

func NewNyLineReader(ed *ny.Editor) *NyLineReader {
	return &NyLineReader{
		ed:   ed,
		hist: simplehistory.New(),
	}
}

func (lr *NyLineReader) SetPrompt(prompt string) {
	lr.ed.Prompt = func() (int, error) {
		return fmt.Fprint(lr.ed.Writer, prompt)
	}
}

func (lr *NyLineReader) ReadLine(ctx context.Context) (string, error) {
	return lr.ed.ReadLine(ctx)
}

func (lr *NyLineReader) Close() error {
	return nil
}

func (lr *NyLineReader) AddToHistory(item string) {
	lr.hist.Add(item)
}

func (sh *Shell) createGohxsLineReader(prompt string) error {
	var styler string
	if !color.NoColor {
		styler = os.Getenv(envStyler)
		if styler == "" {
			styler = "clc-default"
		}
	}
	formatter := os.Getenv(envFormatter)
	if formatter == "" || !strings.HasPrefix(formatter, "terminal") {
		formatter = "terminal"
	}
	cmdColor := color.New(color.Bold)
	cfg := &gohxs.Config{
		Prompt:          prompt,
		HistoryFile:     sh.historyPath,
		InterruptPrompt: "^C",
		EOFPrompt:       `\exit`,
		Output: func(input string) string {
			if strings.HasPrefix(input, CmdPrefix) {
				return cmdColor.Sprint(input)
			}
			buf := bytes.NewBuffer([]byte{})
			err := quick.Highlight(buf, input, "sql", formatter, styler)
			if err != nil {
				log.Fatal(err)
			}
			return buf.String()
		},
		HistorySearchFold: true,
		Stdout:            sh.stdout,
		Stderr:            sh.stderr,
	}
	if sh.historyPath != "" {
		cfg.HistoryFile = sh.historyPath
	}
	rl, err := gohxs.NewEx(cfg)
	if err != nil {
		return err
	}
	sh.lr = NewGohxsLineReader(rl)
	return nil
}

type GohxsLineReader struct {
	rl *gohxs.Instance
}

func NewGohxsLineReader(rl *gohxs.Instance) *GohxsLineReader {
	return &GohxsLineReader{rl: rl}
}

func (lr *GohxsLineReader) SetPrompt(prompt string) {
	lr.rl.SetPrompt(prompt)
}

func (lr *GohxsLineReader) ReadLine(context.Context) (string, error) {
	return lr.rl.Readline()
}

func (lr *GohxsLineReader) Close() error {
	return lr.rl.Close()
}

func (lr *GohxsLineReader) AddToHistory(string) {
	// pass
}
