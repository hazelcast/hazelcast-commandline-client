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
	long := `Outputs the logs of the given Viridian cluster as a stream.

Make sure you authenticate to the Viridian API using 'viridian login' before running this command.

The log format may be one of:
	* minimal: Only the log message
	* basic: Time, level and the log message
	* detailed: Time, level, thread, logger and the log message
	* free form template, see: https://pkg.go.dev/text/template for the format.
	You can use the following placeholders: msg, level, time, thread and logger.
`
	short := "Streams logs of a Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(propLogFormat, "", "basic", false,
		"set the log format, either predefined or free form")
	return nil
}

func (cm StreamLogCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	f := ec.Props().GetString(propLogFormat)
	t, err := template.New("log").Parse(loggerTemplate(f))
	if err != nil {
		return fmt.Errorf("invalid log format %s: %w", f, err)
	}
	clusterNameOrID := ec.Args()[0]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		if s, ok := sp.(clc.SpinnerPauser); ok {
			s.Pause()
		}
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
			kvs["level"] = "DEBUG"
			kvs["time"] = "N/A"
			kvs["thread"] = "N/A"
			kvs["logger"] = "N/A"
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

func loggerTemplate(format string) string {
	switch format {
	case "minimal":
		return "{{.msg}}"
	case "basic", "":
		return "{{.time}} [{{.level}}] {{.msg}}"
	case "detailed":
		return "{{.time}} [{{.level}}] [{{.thread}}] [{{.logger}}] {{.msg}}"
	}
	return format
}

func init() {
	check.Must(plug.Registry.RegisterCommand("viridian:stream-logs", &StreamLogCmd{}))
}
