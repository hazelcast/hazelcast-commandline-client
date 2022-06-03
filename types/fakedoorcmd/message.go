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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	messageFmt   = "The support for %s hasn't been implemented yet.\n\nIf you would like us to implement it, please drop by at:\n%v and add a thumbs up %s.\nWe're happy to implement it quickly based on demand!"
	IssueURLFmt  = "https://github.com/hazelcast/hazelcast-commandline-client/issues/%d"
	thumbsUpSign = "\U0001F44D"
)

type FakeDoor struct {
	Name     string
	IssueNum int
}

func NewFakeCommand(fd FakeDoor) *cobra.Command {
	return &cobra.Command{
		Use:   strings.ToLower(fd.Name),
		Short: fmt.Sprintf("%s operations", fd.Name),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(newFakeDoorMessage(fd))
		},
	}
}

func newFakeDoorMessage(m FakeDoor) string {
	issueNum := fmt.Sprintf(IssueURLFmt, m.IssueNum)
	return fmt.Sprintf(messageFmt, m.Name, issueNum, thumbsUpSign)
}
