package output

const (
	NameKey       = "__key"
	NameKeyType   = "__key_type"
	NameValue     = "this"
	NameValueType = "this_type"
)

type Type int

const (
	TypeDelimited Type = iota
	TypeCSV
	TypeTable
	TypeJSON
)

type SingleLiner interface {
	SingleLine() string
}

type MultiLiner interface {
	MultiLine() []string
}

type RowExtender interface {
	RowExtensions() []Column
}

type Row []Column

func (r Row) MarshalJSON() ([]byte, error) {
	return []byte(`{"fdo": 42}`), nil
}

type RowProducer interface {
	NextRow() (Row, bool)
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

func (s *SimpleRows) NextRow() (Row, bool) {
	if s.index >= len(s.rows) {
		return nil, false
	}
	row := s.rows[s.index]
	s.index++
	return row, true
}

func MaterializeRows(rp RowProducer) []Row {
	var rs []Row
	for {
		row, ok := rp.NextRow()
		if !ok {
			return rs
		}
		rs = append(rs, row)
	}
}
