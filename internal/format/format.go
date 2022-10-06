package format

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

// Fmt defines output format for different SQL types
func Fmt(v interface{}) string {
	switch t := v.(type) {
	case types.LocalDate:
		return time.Time(t).Format("2006-01-02")
	case types.LocalDateTime:
		return time.Time(t).Format("2006-01-02T15:04:05.999999")
	case types.LocalTime:
		return time.Time(t).Format("15:04:05.999999")
	case types.OffsetDateTime:
		return time.Time(t).Format(time.RFC3339)
	default:
		return fmt.Sprint(t)
	}
}

const (
	Pretty = "pretty"
	CSV    = "csv"
	JSON   = "json"
)

func ValidFormats() []string {
	return []string{Pretty, CSV, JSON}
}

type Writer func(values ...interface{}) error

type writerBuilder struct {
	format  string
	headers []interface{}
	values  [][]interface{}
	out     io.Writer
}

func NewWriterBuilder() *writerBuilder {
	return &writerBuilder{
		format: Pretty,
		out:    os.Stderr,
	}
}

func (wb *writerBuilder) WithFormat(f string) *writerBuilder {
	wb.format = f
	return wb
}

func (wb *writerBuilder) WithHeaders(headers ...interface{}) *writerBuilder {
	wb.headers = headers
	return wb
}

func (wb *writerBuilder) WithOut(out io.Writer) *writerBuilder {
	wb.out = out
	return wb
}

func IsValid(f string) bool {
	for _, e := range ValidFormats() {
		if f == e {
			return true
		}
	}
	return false
}

func (wb *writerBuilder) Build() (Writer, error) {
	if !IsValid(wb.format) {
		return nil, errors.New("unknown format")
	}
	var valueWriter, headerWriter Writer
	headersSet := wb.headers != nil
	switch wb.format {
	case Pretty:
		valueWriter, headerWriter = prettyWriter(wb.out)
	case CSV:
		w := csv.NewWriter(wb.out)
		headerWriter, valueWriter = csvWriter(w), csvWriter(w)
	case JSON:
		// assign nil writer for headers
		headerWriter = func(_ ...interface{}) error {
			return nil
		}
		enc := json.NewEncoder(wb.out)
		enc.SetIndent("", "")
		valueWriter = jsonWriter(wb.out, wb.headers)
	default:
		return nil, errors.New("unkown output format")
	}
	if headersSet {
		if err := headerWriter(wb.headers...); err != nil {
			return nil, err
		}
		valueWriter = valueLengthWrapperWriter(len(wb.headers), valueWriter)
	}
	return valueWriter, nil
}

func valueLengthWrapperWriter(headerLen int, w Writer) Writer {
	return func(values ...interface{}) error {
		if headerLen != len(values) {
			return errors.New("element and header counts must match")
		}
		return w(values...)
	}
}

func PrintValueBasedOnType(cmd *cobra.Command, value interface{}) {
	var err error
	switch v := value.(type) {
	case serialization.JSON:
		if err = quick.Highlight(cmd.OutOrStdout(), fmt.Sprintln(v),
			"json", "terminal", "tango"); err != nil {
			cmd.Println(v.String())
		}
	default:
		if v == nil {
			cmd.Println("There is no value corresponding to the provided key")
			break
		}
		cmd.Println(v)
	}
}

func prettyWriter(out io.Writer) (valueWriter Writer, headerWriter Writer) {
	w := table.NewTableWriter(out)
	headerWriter = w.WriteHeader
	valueWriter = func(values ...interface{}) error {
		return w.Write(values...)
	}
	return valueWriter, headerWriter
}

func csvWriter(w *csv.Writer) func(values ...interface{}) error {
	return func(values ...interface{}) error {
		sValues := make([]string, len(values))
		for i, v := range values {
			sValues[i] = fmt.Sprint(v)
		}
		if err := w.Write(sValues); err != nil {
			return err
		}
		w.Flush()
		return nil
	}
}

func jsonWriter(out io.Writer, headers []interface{}) Writer {
	headersSet := len(headers) != 0
	var bb bytes.Buffer
	return func(values ...interface{}) error {
		m := make(map[string]interface{})
		for i, v := range values {
			if headersSet {
				m[fmt.Sprint(headers[i])] = v
			} else {
				m[strconv.Itoa(i)] = v
			}
		}
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = json.Compact(&bb, b)
		if err != nil {
			return err
		}
		_, err = bb.WriteTo(out)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, "")
		return err
	}
}
