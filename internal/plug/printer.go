package plug

import (
	"context"
	"io"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

type Printer interface {
	PrintStream(ctx context.Context, w io.Writer, rp output.RowProducer) error
	PrintRows(ctx context.Context, w io.Writer, rows []output.Row) error
}
