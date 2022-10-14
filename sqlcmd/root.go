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
	"github.com/hazelcast/hazelcast-commandline-client/internal/connection"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
)

func New(config *hazelcast.Config) *cobra.Command {
	//var outputType string
	cmd := &cobra.Command{
		Use:     "sql [query]",
		Short:   "Execute the given SQL query",
		Example: `sql 'select * from table(generate_stream(1))'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ot, err := output.TypeStringFor(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			c, err := connection.ConnectToCluster(ctx, config)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Cannot get initialize SQL driver")
			}
			q := strings.Join(args, " ")
			q = strings.TrimSpace(q)
			if len(q) == 0 {
				return hzcerrors.NewLoggableError(err, "Query is required")
			}
			// If a statement is provided, run it in non-interactive mode
			lt := strings.ToLower(q)
			if strings.HasPrefix(lt, "select") || strings.HasPrefix(lt, "show") {
				if err := query(ctx, c, q, cmd.OutOrStdout(), ot); err != nil && !isContextCancellationErr(err) {
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
	//decorateCommandWithOutputFlag(&outputType, cmd)
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

//
//func decorateCommandWithOutputFlag(outputType *string, cmd *cobra.Command) {
//	flags := cmd.Flags()
//	opts := []string{outputDefault, outputPretty, outputCSV, outputJSON}
//	flags.StringVarP(outputType, "output-type", "o", outputPretty, fmt.Sprintf("one of: %s", strings.Join(opts, ", ")))
//	cmd.RegisterFlagCompletionFunc("output-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
//		return []string{outputPretty, outputCSV}, cobra.ShellCompDirectiveDefault
//	})
//}
