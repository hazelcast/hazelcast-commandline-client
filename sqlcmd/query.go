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
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal/format"
	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

func query(ctx context.Context, c *hazelcast.Client, text string, out io.Writer, outputType string) error {
	result, err := c.SQL().Execute(ctx, text)
	if err != nil {
		return fmt.Errorf("querying: %w", err)
	}
	defer func() {
		ch := make(chan struct{})
		go func() {
			result.Close()
			close(ch)
		}()
		// result.Close blocks if there are no members to communicate with, so do not wait more than 2 secs.
		select {
		case <-time.After(2 * time.Second):
		case <-ch:
		}
	}()
	switch outputType {
	case outputPretty:
		tWriter := table.NewTableWriter(out)
		return rowsHandler(result, func(cols []string) error {
			icols := make([]interface{}, len(cols))
			for i, v := range cols {
				icols[i] = v
			}
			return tWriter.WriteHeader(icols...)
		}, func(row []interface{}) error {
			for i, v := range row {
				row[i] = format.Fmt(v)
			}
			return tWriter.Write(row...)
		})
	case outputCSV:
		csvWriter := csv.NewWriter(out)
		return rowsHandler(result, func(cols []string) error {
			if err := csvWriter.Write(cols); err != nil {
				return err
			}
			csvWriter.Flush()
			return nil
		}, func(values []interface{}) error {
			strValues := make([]string, len(values))
			for i, v := range values {
				strValues[i] = format.Fmt(v)
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
func rowsHandler(result sql.Result, columnHandler func(cols []string) error, rowHandler func([]interface{}) error) error {
	mt, err := result.RowMetadata()
	if err != nil {
		return fmt.Errorf("retrieving metadata: %w", err)
	}
	var cols []string
	for _, c := range mt.Columns() {
		cols = append(cols, c.Name())
	}
	if err = columnHandler(cols); err != nil {
		return err
	}
	it, err := result.Iterator()
	if err != nil {
		return fmt.Errorf("initializing result iterator: %w", err)
	}
	for it.HasNext() {
		row, err := it.Next()
		if err != nil {
			return fmt.Errorf("fetching row: %w", err)
		}
		values := make([]interface{}, len(cols))
		for i := 0; i < len(cols); i++ {
			v, err := row.Get(i)
			if err != nil {
				return fmt.Errorf("fetching value: %w", err)
			}
			values[i] = v
		}
		if err := rowHandler(values); err != nil {
			return err
		}
	}
	return nil
}

func execute(ctx context.Context, c *hazelcast.Client, text string) error {
	r, err := c.SQL().Execute(ctx, text)
	if err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	uc := r.UpdateCount()
	if uc == -1 {
		return fmt.Errorf("invalid update count")
	}
	cmd.Printf("---\nAffected rows: %d\n\n", uc)
	return nil
}
