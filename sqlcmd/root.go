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

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/browser"
	"github.com/hazelcast/hazelcast-commandline-client/internal/tuiutil"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
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
			//todo create driver from existing client
			driver, err := internal.SQLDriver(ctx, config)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Cannot get initialize SQL driver")
			}
			q := strings.Join(args, " ")
			q = strings.TrimSpace(q)
			if len(q) == 0 {
				// If no queries given, run sql browser
				p := browser.InitSQLBrowser(driver, tuiutil.SelectedTheme == tuiutil.NoColor)
				if err := p.Start(); err != nil {
					fmt.Println("could not run sql browser:", err)
					return err
				}
				return nil
			}
			// If a statement is provided, run it in non-interactive mode
			lt := strings.ToLower(q)
			if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
				if err := query(ctx, driver, q, cmd.OutOrStdout(), outputType); err != nil {
					return hzcerrors.NewLoggableError(err, "Cannot execute the query")
				}
			} else {
				if err := execute(ctx, driver, q); err != nil {
					return hzcerrors.NewLoggableError(err, "Cannot execute the query")
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
