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
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerror "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

const (
	outputPretty = "pretty"
	outputCSV    = "csv"
)

func NewQuery(config *hazelcast.Config) *cobra.Command {
	var outputType string
	cmd := &cobra.Command{
		Use:   `query statement-string`,
		Short: "Executes SQL query",
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := strings.Join(args, " ")
			var queries []string
			for _, q := range strings.Split(arg, ";") {
				tmp := strings.TrimSpace(q)
				if len(tmp) == 0 {
					continue
				}
				queries = append(queries, tmp)
			}
			if len(queries) == 0 {
				return cmd.Help()
			}
			if outputType != outputPretty && outputType != outputCSV {
				return hzcerror.NewLoggableError(nil,
					"Provided output type parameter (%s) is not a known type. Provide either '%s' or '%s'",
					outputType, outputPretty, outputCSV)
			}
			ctx := cmd.Context()
			c, err := internal.ConnectToCluster(ctx, config)
			if err != nil {
				return hzcerror.NewLoggableError(err, "Cannot get initialize client")
			}
			for _, q := range queries {
				lt := strings.ToLower(q)
				if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
					if err := query(ctx, c, q, cmd.OutOrStdout(), outputType); err != nil {
						return hzcerror.NewLoggableError(err, "Cannot execute the query")
					}
				} else {
					if err := execute(ctx, c, q); err != nil {
						return hzcerror.NewLoggableError(err, "Cannot execute the query")
					}
				}
			}
			return nil
		},
	}
	decorateCommandWithOutputFlag(&outputType, cmd)
	return cmd
}

func decorateCommandWithOutputFlag(outputType *string, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(outputType, "output-type", "o", outputPretty, fmt.Sprintf("%s or %s", outputPretty, outputCSV))
	cmd.RegisterFlagCompletionFunc("output-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{outputPretty, outputCSV}, cobra.ShellCompDirectiveDefault
	})
}

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
