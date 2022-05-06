package flagdecorators

import (
	"github.com/spf13/cobra"
)

const (
	JsonEntryFlag = "json-entry"
	AllEntryFlag  = "all"
)

func DecorateCommandWithAllFlag(cmd *cobra.Command, all *bool, required bool, usage string) {
	cmd.Flags().BoolVar(all, AllEntryFlag, false, usage)
	if required {
		cmd.MarkFlagRequired(AllEntryFlag)
	}
}

func DecorateCommandWithJsonEntryFlag(cmd *cobra.Command, jsonEntry *string, required bool, usage string) {
	cmd.Flags().StringVar(jsonEntry, JsonEntryFlag, "", usage)
	if required {
		cmd.MarkFlagRequired(JsonEntryFlag)
	}
}
