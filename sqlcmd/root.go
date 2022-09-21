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
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/browser"
)

const (
	outputPretty = "pretty"
	outputCSV    = "csv"
)

func New(config *hazelcast.Config) *cobra.Command {
	var outputType string
	cmd := &cobra.Command{
		Use:   "sql [query]",
		Short: "Start SQL Browser or execute given SQL query",
		Example: `sql 	# starts the SQL Browser
sql "CREATE MAPPING IF NOT EXISTS myMap (__key VARCHAR, this VARCHAR) TYPE IMAP OPTIONS ( 'keyFormat' = 'varchar', 'valueFormat' = 'varchar')" 	# executes the query`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputType != outputPretty && outputType != outputCSV {
				return hzcerrors.NewLoggableError(nil,
					"Provided output type parameter (%s) is not a known type. Provide either '%s' or '%s'",
					outputType, outputPretty, outputCSV)
			}
			ctx := cmd.Context()
			c, err := internal.ConnectToCluster(ctx, config)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Cannot get initialize SQL driver")
			}
			q := strings.Join(args, " ")
			q = strings.TrimSpace(q)
			if len(q) == 0 {
				// If no queries given, run sql browser
				p := browser.InitSQLBrowser(c)
				if err := p.Start(); err != nil {
					fmt.Println("could not run sql browser:", err)
					return err
				}
				return nil
			}
			// If a statement is provided, run it in non-interactive mode
			lt := strings.ToLower(q)
			if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
				if err := query(ctx, c, q, cmd.OutOrStdout(), outputType); err != nil && !isContextCancellationErr(err) {
					return hzcerrors.NewLoggableError(err, "Cannot execute the query")
				}
			} else {
				r, err := execute(ctx, c, q)
				if err != nil && !isContextCancellationErr(err) {
					return hzcerrors.NewLoggableError(err, "Cannot execute the query")
				}
				uc := r.UpdateCount()
				if uc == -1 {
					return hzcerrors.NewLoggableError(fmt.Errorf("invalid update count"), "Cannot execute the query")
				}
				cmd.Printf("---\nAffected rows: %d\n\n", uc)
			}
			return nil
		},
	}
	decorateCommandWithOutputFlag(&outputType, cmd)
	return cmd
}

func isContextCancellationErr(err error) bool {
	errTxt := err.Error()
	// todo find a better way to detect user ended the query
	if strings.Contains(errTxt, "context canceled") {
		// do not print error for context cancellation.
		return true
	}
	return false
}

func decorateCommandWithOutputFlag(outputType *string, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(outputType, "output-type", "o", outputPretty, fmt.Sprintf("%s or %s", outputPretty, outputCSV))
	cmd.RegisterFlagCompletionFunc("output-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{outputPretty, outputCSV}, cobra.ShellCompDirectiveDefault
	})
}
