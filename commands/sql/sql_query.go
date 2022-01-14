/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/alexeyco/simpletable"
	_ "github.com/hazelcast/hazelcast-go-client/sql/driver"
	"github.com/spf13/cobra"

	isql "github.com/hazelcast/hazelcast-commandline-client/internal/sql"
)

var queryCmd = &cobra.Command{
	Use:   `query string`,
	Short: "executes query",
	Run: func(cmd *cobra.Command, args []string) {
		var queries []string
		for _, arg := range args {
			for _, q := range strings.Split(arg, ";") {
				tmp := strings.TrimSpace(q)
				if len(tmp) == 0 {
					continue
				}
				queries = append(queries, tmp)
			}
		}
		db := isql.Get()
		for _, q := range queries {
			fmt.Println(">>>", q)
			lt := strings.ToLower(q)
			if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
				if err := query(db, q); err != nil {
					fmt.Println(err)
				}
			} else {
				if err := exec(db, q); err != nil {
					fmt.Println(err)
				}
			}
		}
		return
	},
}

func query(db *sql.DB, text string) error {
	rows, err := db.Query(text)
	if err != nil {
		return fmt.Errorf("querying: %w", err)
	}
	defer rows.Close()
	// Generate table from rows
	table := simpletable.New()
	var header []*simpletable.Cell
	err = rowsHandler(rows, func(cols []string) error {
		for _, col := range cols {
			header = append(header, &simpletable.Cell{
				Align: simpletable.AlignCenter,
				Text:  col,
			})
		}
		return nil
	}, func(row []interface{}) error {
		cells := make([]*simpletable.Cell, len(row))
		for i, e := range row {
			cells[i] = &simpletable.Cell{
				Align: simpletable.AlignLeft,
				// TODO we can create formatters for each sql type
				Text: fmt.Sprintf("%v", e),
			}
		}
		table.Body.Cells = append(table.Body.Cells, cells)
		return nil
	})
	if err != nil {
		fmt.Println("could not generate table: ", err)
		return nil
	}
	table.Header = &simpletable.Header{Cells: header}
	table.Footer = &simpletable.Footer{}
	table.Footer.Cells = make([]*simpletable.Cell, len(table.Header.Cells))
	footer := table.Footer.Cells
	for i := range footer {
		footer[i] = &simpletable.Cell{}
	}
	footer[0] = &simpletable.Cell{Align: simpletable.AlignRight, Text: fmt.Sprintf("#Rows:%d", len(table.Body.Cells))}
	table.SetStyle(simpletable.StyleCompactLite)
	table.Println()
	return nil
}

// Reads columns and rows calls handlers. rowHandler is called per row.
func rowsHandler(rows *sql.Rows, columnHandler func(cols []string) error, rowHandler func([]interface{}) error) error {
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("retrieving columns: %w", err)
	}
	if err = columnHandler(cols); err != nil {
		return err
	}
	emptyRow := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		emptyRow[i] = new(interface{})
	}
	for rows.Next() {
		row := make([]interface{}, len(emptyRow))
		copy(row, emptyRow)
		if err := rows.Scan(row...); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}
		for i := range row {
			row[i] = *(row[i].(*interface{}))
		}
		if err := rowHandler(row); err != nil {
			return err
		}
	}
	return nil
}

func exec(db *sql.DB, text string) error {
	r, err := db.Exec(text)
	if err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	ra, err := r.RowsAffected()
	if err != nil {
		return nil
	}
	fmt.Printf("---\nAffected rows: %d\n\n", ra)
	return nil
}
