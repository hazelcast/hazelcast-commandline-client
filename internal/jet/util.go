package jet

import (
	"bytes"
	"crypto/sha256"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultBatchSize                          = 2 * 1024 * 1024 // 10MB
	JobStatusNotRunning                       = 0
	JobStatusStarting                         = 1
	JobStatusRunning                          = 2
	JobStatusSuspended                        = 3
	JobStatusSuspendedExportingSnapshot       = 4
	JobStatusCompleting                       = 5
	JobStatusFailed                           = 6
	JobStatusCompleted                        = 7
	TerminateModeRestartGraceful        int32 = 0
	TerminateModeRestartForceful        int32 = 1
	TerminateModeSuspendGraceful        int32 = 2
	TerminateModeSuspendForceful        int32 = 3
	TerminateModeCancelGraceful         int32 = 4
	TerminateModeCancelForceful         int32 = 5
)

func hashOfPath(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return calculateHash(f)
}

func calculateHash(r io.Reader) ([]byte, error) {
	h := sha256.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, 32)
	b := h.Sum(buf)
	return b, nil
}

func partCountOf(path string, partSize int) (int, error) {
	st, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(math.Ceil(float64(st.Size()) / float64(partSize))), nil
}

func StatusToString(status int32) string {
	switch status {
	case JobStatusNotRunning:
		return "NOT_RUNNING"
	case JobStatusStarting:
		return "STARTING"
	case JobStatusRunning:
		return "RUNNING"
	case JobStatusSuspended:
		return "SUSPENDED"
	case JobStatusSuspendedExportingSnapshot:
		return "SUSPENDED_EXPORTING_SNAPSHOT"
	case JobStatusCompleting:
		return "COMPLETING"
	case JobStatusFailed:
		return "FAILED"
	case JobStatusCompleted:
		return "COMPLETED"
	}
	return "UNKNOWN"
}

type binBatch struct {
	reader io.Reader
	buf    []byte
}

func newBatch(reader io.Reader, batchSize int) *binBatch {
	if batchSize < 1 {
		panic("newBatch: batchSize must be positive")
	}
	return &binBatch{
		reader: reader,
		buf:    make([]byte, batchSize),
	}
}

// Next returns the next batch of bytes.
// Make sure to copy it before calling Next again.
func (bb *binBatch) Next() ([]byte, []byte, error) {
	n, err := bb.reader.Read(bb.buf)
	if err != nil {
		return nil, nil, err
	}
	b := bb.buf[0:n:n]
	h, err := calculateHash(bytes.NewBuffer(b))
	if err != nil {
		return nil, nil, err
	}
	return b, h, nil
}

type PathBinaryReader struct {
	path string
}

func CreateBinaryReaderForPath(path string) PathBinaryReader {
	return PathBinaryReader{path: path}
}

func (pch PathBinaryReader) Hash() ([]byte, error) {
	f, err := os.Open(pch.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return calculateHash(f)
}

func (pch PathBinaryReader) Reader() (io.ReadCloser, error) {
	return os.Open(pch.path)
}

func (pch PathBinaryReader) FileName() string {
	_, fn := filepath.Split(pch.path)
	return strings.TrimSuffix(fn, ".jar")
}

func (pch PathBinaryReader) PartCount(partSize int) (int, error) {
	st, err := os.Stat(pch.path)
	if err != nil {
		return 0, err
	}
	return int(math.Ceil(float64(st.Size()) / float64(partSize))), nil
}
