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
	"errors"
	"fmt"
	"strings"
	"syscall"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/browser"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
)

func New(config *hazelcast.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "sql [query]",
		Short: "Start SQL Browser or execute given SQL query",
		Example: `sql 	# starts the SQL Browser
sql "CREATE MAPPING IF NOT EXISTS myMap (__key VARCHAR, this VARCHAR) TYPE IMAP OPTIONS ( 'keyFormat' = 'varchar', 'valueFormat' = 'varchar')" 	# executes the query`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := internal.ConnectToCluster(ctx, config)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Cannot get initialize client")
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
				if err := query(ctx, c, q, cmd.OutOrStdout(), true); err != nil {
					if errors.Is(err, syscall.EPIPE) {
						// pager may be closed, expected error
						return nil
					}
					return hzcerrors.NewLoggableError(err, "Cannot execute the query")
				}
			} else {
				if err := execute(ctx, c, q); err != nil {
					return hzcerrors.NewLoggableError(err, "Cannot execute the query")
				}
			}
			return nil
		},
	}
	return &cmd
}
