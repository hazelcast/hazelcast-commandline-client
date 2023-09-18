//go:build std || viridian

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
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	propLogFormat = "log-format"
)

type StreamLogCommand struct{}

func (StreamLogCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("stream-logs")
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
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(propLogFormat, "", "basic", false,
		"set the log format, either predefined or free form")
	cc.AddStringArg(argClusterID, argTitleClusterID)
	return nil
}

func (StreamLogCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.Metrics().Increment(metric.NewSimpleKey(), "total.viridian."+cmd.RunningMode(ec))
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	f := ec.Props().GetString(propLogFormat)
	t, err := template.New("log").Parse(loggerTemplate(f))
	if err != nil {
		return fmt.Errorf("invalid log format %s: %w", f, err)
	}
	clusterNameOrID := ec.GetStringArg(argClusterID)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		lf := newLogFixer(ec.Stdout(), t)
		for {
			if err = api.StreamLogs(ctx, clusterNameOrID, lf); err != nil {
				if err.Error() == "unexpected EOF" {
					// disconnected, advance the log time so observed log lines aren't displayed again
					// then reconnect
					lf.Advance()
					continue
				}
				return nil, err
			}
			break
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
	buf          *bytes.Buffer
	inner        io.Writer
	tmpl         *template.Template
	lastTime     time.Time
	nextLastTime time.Time
}

func newLogFixer(wrapped io.Writer, tmpl *template.Template) *logFixer {
	return &logFixer{
		buf:   &bytes.Buffer{},
		inner: wrapped,
		tmpl:  tmpl,
	}
}

func (lf *logFixer) Advance() {
	lf.lastTime = lf.nextLastTime
}

func (lf *logFixer) Write(p []byte) (int, error) {
	n, err := lf.buf.Write(p)
	if err != nil {
		return 0, fmt.Errorf("logFixer.Write: writing to buffer")
	}
	kvs := map[string]any{}
	scn := bufio.NewScanner(lf.buf)
	// using a custom splitter to keep CR at the end of the line
	scn.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			// We have a full newline-terminated line.
			return i + 1, data[0 : i+1], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	})
	var nullTime time.Time
	for scn.Scan() {
		if scn.Err() != nil {
			return 0, fmt.Errorf("logFixer.Write: scanning: %w", scn.Err())
		}
		line := scn.Text()
		if !strings.HasSuffix(line, "\n") {
			// this is the last part of the input
			// write it back and quit
			lf.buf.Write(scn.Bytes())
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = line[5:]
		}
		line = strings.TrimSuffix(line, "\n")
		if line == "" {
			continue
		}
		var logTime time.Time
		if err = json.Unmarshal([]byte(line), &kvs); err != nil {
			kvs["msg"] = line
			kvs["level"] = "DEBUG"
			kvs["time"] = "N/A"
			kvs["thread"] = "N/A"
			kvs["logger"] = "N/A"
		} else {
			// try to convert the time
			if v, ok := kvs["time"]; ok {
				t, err := time.Parse(time.RFC3339, v.(string))
				// note: checking err == nil
				if err == nil {
					logTime = t
				}
			}
		}
		if logTime.After(nullTime) {
			// time for the log line was calculated
			if !logTime.After(lf.lastTime) {
				// this log line was already observed
				continue
			}
			// this is a new log line
			// activate it in the next call to the API
			lf.nextLastTime = logTime
		} else if lf.lastTime.After(nullTime) {
			// time for the log line was not calculated
			// this is probably an old log line
			continue
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
	check.Must(plug.Registry.RegisterCommand("viridian:stream-logs", &StreamLogCommand{}))
}
