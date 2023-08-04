package output

import "context"

const (
	NameKey       = "__key"
	NameKeyType   = "__key_type"
	NameValue     = "this"
	NameValueType = "this_type"
)

type Type int

type RowExtender interface {
	RowExtensions() []Column
}

type Row []Column

type RowProducer interface {
	NextRow(ctx context.Context) (Row, bool, error)
}

type RowConsumer interface {
	AddRow(row Row)
}

type SimpleRows struct {
	rows  []Row
	index int
}

func NewSimpleRows(rows []Row) *SimpleRows {
	return &SimpleRows{rows: rows}
}

func (s *SimpleRows) NextRow(ctx context.Context) (Row, bool, error) {
	if s.index >= len(s.rows) {
		return nil, false, nil
	}
	row := s.rows[s.index]
	s.index++
	return row, true, nil
}

type ChanRows struct {
	ch <-chan Row
}

func NewChanRows(ch <-chan Row) *ChanRows {
	return &ChanRows{ch: ch}
}

func (c ChanRows) NextRow(ctx context.Context) (Row, bool, error) {
	select {
	case row, ok := <-c.ch:
		return row, ok, nil
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
}
