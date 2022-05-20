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

package sqlcmd

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

func query(ctx context.Context, c *hazelcast.Client, text string, out io.Writer, outputType string) error {
	rows, err := c.QuerySQL(ctx, text)
	if err != nil {
		return fmt.Errorf("querying: %w", err)
	}
	defer rows.Close()
	switch outputType {
	case outputPretty:
		tWriter := table.NewTableWriter(out)
		return rowsHandler(rows, func(cols []string) error {
			icols := make([]interface{}, len(cols))
			for i, v := range cols {
				icols[i] = v
			}
			return tWriter.WriteHeader(icols...)
		}, func(row []interface{}) error {
			return tWriter.Write(row...)
		})
	case outputCSV:
		csvWriter := csv.NewWriter(out)
		return rowsHandler(rows, func(cols []string) error {
			if err := csvWriter.Write(cols); err != nil {
				return err
			}
			csvWriter.Flush()
			return nil
		}, func(values []interface{}) error {
			strValues := make([]string, len(values))
			for i, v := range values {
				strValues[i] = fmt.Sprint(v)
			}
			if err := csvWriter.Write(strValues); err != nil {
				return err
			}
			csvWriter.Flush()
			return nil
		})
	}
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

func execute(ctx context.Context, c *hazelcast.Client, text string) error {
	r, err := c.ExecSQL(ctx, text)
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
