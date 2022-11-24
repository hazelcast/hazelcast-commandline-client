package plug

import (
	"io"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

type Printer interface {
	Print(w io.Writer, rp output.RowProducer) error
}
