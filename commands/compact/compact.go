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
package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal/compact"
)

var language string
var schemaFilePath string
var outputDir string
var namespace string
var CompactCmd = &cobra.Command{
	Use: "compact LANGUAGE SCHEMA_FILE [flags]\n" +
		"\nArguments:\n" +
		"  LANGUAGE    Language to generate codecs for. Possible values are [java cpp cs py ts go]\n" +
		"  SCHEMA_FILE Path to the schema file.\n",
	Short: "Hazelcast Code Generator for Compact Serializer",
	Long:  "Generate domain classes from given schema file for the selected language.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("LANGUAGE and SCHEMA_FILE are required")
		}
		if len(args) > 3 {
			return errors.New("too many arguments after SCHEMA_FILE")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		language = args[0]
		schemaFilePath = args[1]
		fmt.Printf(">>>>> language %s, schemaFilePath %s, OutputDir %s, namespace %s\n", language, schemaFilePath, outputDir, namespace)

		err := compact.Generate(language, schemaFilePath, outputDir, namespace)
		if err == nil {
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	CompactCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Output directory")
	CompactCmd.Flags().StringVar(&namespace, "namespace", "", "Namespace of the classes generated. Namespace interpreted differently according to selected language.\njava: namespace is used as the package name of the classes")
}
