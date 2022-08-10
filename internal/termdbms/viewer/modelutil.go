package viewer

import (
	"database/sql"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *TuiModel) WriteMessage(s string) {
	if Message == "" {
		Message = s
		MIP = true
		go Program.Send(tea.KeyMsg{}) // trigger update
		go Program.Send(tea.KeyMsg{}) // trigger update for sure hack gross but w/e
	}
}

func (m *TuiModel) CopyMap() (to map[string]interface{}) {
	from := m.Table().Data
	to = map[string]interface{}{}

	for k, v := range from {
		if copyValues, ok := v.(map[string][]interface{}); ok {
			columnNames := m.Data().TableHeaders[k]
			columnValues := make(map[string][]interface{})
			// golang wizardry
			columns := make([]interface{}, len(columnNames))

			for i := range columns {
				columns[i] = copyValues[columnNames[i]]
			}

			for i, colName := range columnNames {
				val := columns[i].([]interface{})
				buffer := make([]interface{}, len(val))
				for k := range val {
					buffer[k] = val[k]
				}
				columnValues[colName] = append(columnValues[colName], buffer)
			}

			to[k] = columnValues // data for schema, organized by column
		}
	}

	return to
}

// GetNewModel returns a TuiModel struct with some fields set
func GetNewModel(_ string, _ *sql.DB) TuiModel {
	m := TuiModel{
		DefaultTable: TableState{
			Data: make(map[string]interface{}),
		},
		Format: FormatState{
			EditSlices:     nil,
			Text:           nil,
			RunningOffsets: nil,
			CursorX:        0,
			CursorY:        0,
		},
		UI: UIState{
			CanFormatScroll:   false,
			RenderSelection:   false,
			EditModeEnabled:   false,
			FormatModeEnabled: false,
			BorderToggle:      false,
			CurrentTable:      0,
			ExpandColumn:      -1,
		},
		Scroll: ScrollData{},
		DefaultData: UIData{
			TableHeaders:      make(map[string][]string),
			TableHeadersSlice: []string{},
			TableSlices:       make(map[string][]interface{}),
			TableIndexMap:     make(map[int]string),
		},
	}
	return m
}

// SetModel creates a model to be used by bubbletea using some golang wizardry
func (m *TuiModel) SetModel(c *sql.Rows, db *sql.DB) error {
	var err error
	indexMap := 0
	// gets all the schema names of the database
	tableNamesQuery := m.Table().Database.GetTableNamesQuery()
	rows, err := db.Query(tableNamesQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	// for each schema
	for rows.Next() {
		var schemaName string
		rows.Scan(&schemaName)
		// couldn't get prepared statements working and gave up because it was very simple
		var statement strings.Builder
		statement.WriteString("select * from ")
		statement.WriteString(schemaName)
		getAll := statement.String()
		if c != nil {
			c.Close()
			c = nil
		}
		c, err = db.Query(getAll)
		if err != nil {
			panic(err)
		}
		m.PopulateDataForResult(c, &indexMap, schemaName)
	}
	// set the first table to be initial view
	m.UI.CurrentTable = 1
	return nil
}

func (m *TuiModel) PopulateDataForResult(c *sql.Rows, indexMap *int, schemaName string) {
	columnNames, _ := c.Columns()
	columnValues := make(map[string][]interface{})
	for c.Next() { // each row of the table
		// golang wizardry
		columns := make([]interface{}, len(columnNames))
		columnPointers := make([]interface{}, len(columnNames))
		// init interface array
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		c.Scan(columnPointers...)
		for i, colName := range columnNames {
			val := columnPointers[i].(*interface{})
			columnValues[colName] = append(columnValues[colName], *val)
		}
	}
	// onto the next schema
	*indexMap++
	if m.QueryResult != nil && m.QueryData != nil {
		m.QueryResult.Data[schemaName] = columnValues
		m.QueryData.TableHeaders[schemaName] = columnNames // headers for the schema, for later reference
		m.QueryData.TableIndexMap[*indexMap] = schemaName
		return
	}
	m.Table().Data[schemaName] = columnValues       // data for schema, organized by column
	m.Data().TableHeaders[schemaName] = columnNames // headers for the schema, for later reference
	// mapping between schema and an int ( since maps aren't deterministic), for later reference
	m.Data().TableIndexMap[*indexMap] = schemaName
}

func (m *TuiModel) SwapTableValues(f, t *TableState) {
	from := &f.Data
	to := &t.Data
	for k, v := range *from {
		if copyValues, ok := v.(map[string][]interface{}); ok {
			columnNames := m.Data().TableHeaders[k]
			columnValues := make(map[string][]interface{})
			// golang wizardry
			columns := make([]interface{}, len(columnNames))
			for i := range columns {
				columns[i] = copyValues[columnNames[i]][0]
			}
			for i, colName := range columnNames {
				columnValues[colName] = columns[i].([]interface{})
			}
			(*to)[k] = columnValues // data for schema, organized by column
		}
	}
}
