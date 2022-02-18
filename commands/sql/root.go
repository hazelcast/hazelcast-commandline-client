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
	"github.com/spf13/cobra"
)

var SqlCmd = &cobra.Command{
	Use:   "sql {query}",
	Short: "sql operations",
	Example: `sql query "CREATE MAPPING IF NOT EXISTS myMap (__key VARCHAR,val VARCHAR) TYPE IMAP OPTIONS ( 'keyFormat' = 'varchar', 'valueFormat' = 'varchar')"
sql query "select * from myMap"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	SqlCmd.AddCommand(queryCmd)
}
