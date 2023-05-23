package viridian

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	propLogFormat = "log-format"
)

type StreamLogCmd struct{}

func (cm StreamLogCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("stream-logs [cluster-ID/name]")
	long := `Streams logs of the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Streams logs of a Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(propLogFormat, "", "{{.level}} {{.time}}: {{.msg}}", false, "Set the log format")
	return nil
}

func (cm StreamLogCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	f := ec.Props().GetString(propLogFormat)
	t, err := template.New("log").Parse(f)
	if err != nil {
		return fmt.Errorf("invalid log format %s: %w", f, err)
	}
	clusterNameOrID := ec.Args()[0]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		if s, ok := sp.(clc.SpinnerPauser); ok {
			s.Pause()
		}
		//sp.SetText("Streaming the logs")
		lf := newLogFixer(ec.Stdout(), t)
		if err = api.StreamLogs(ctx, clusterNameOrID, lf); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	return nil
}

type logFixer struct {
	buf   *bytes.Buffer
	inner io.Writer
	tmpl  *template.Template
}

func newLogFixer(wrapped io.Writer, tmpl *template.Template) *logFixer {
	return &logFixer{
		buf:   &bytes.Buffer{},
		inner: wrapped,
		tmpl:  tmpl,
	}
}

func (lf *logFixer) Write(p []byte) (int, error) {
	n, err := lf.buf.Write(p)
	if err != nil {
		return 0, fmt.Errorf("logFixer.Write: writing to buffer")
	}
	kvs := map[string]any{}
	scn := bufio.NewScanner(lf.buf)
	for scn.Scan() {
		if scn.Err() != nil {
			return 0, fmt.Errorf("logFixer.Write: scanning: %w", scn.Err())
		}
		line := scn.Text()
		if strings.HasPrefix(line, "data:") {
			line = line[5:]
		}
		if line == "" {
			continue
		}
		if err = json.Unmarshal([]byte(line), &kvs); err != nil {
			kvs["msg"] = line
			kvs["level"] = "INFO"
			kvs["time"] = "UNKNOWN"
			kvs["thread"] = "UNKNOWN"
			kvs["logger"] = "UNKNOWN"
		}
		if err = lf.tmpl.Execute(lf.inner, kvs); err != nil {
			return 0, fmt.Errorf("logFixer.Write: writing: %w", err)
		}
		if _, err = lf.inner.Write([]byte{'\n'}); err != nil {
			return 0, fmt.Errorf("logFixer.Write: writing: %w", err)
		}
	}
	return n, nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:stream-logs", &StreamLogCmd{}))
}
