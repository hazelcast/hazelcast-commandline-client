package flagdecorators

import (
	"time"

	"github.com/spf13/cobra"
)

// common flags
const (
	JsonEntryFlag = "json-entry"
	TTLFlag       = "ttl"
	MaxIdleFlag   = "max-idle"
	DelimiterFlag = "delim"
)

func DecorateCommandWithJsonEntryFlag(cmd *cobra.Command, jsonEntry *string, required bool, usage string) error {
	cmd.Flags().StringVar(jsonEntry, JsonEntryFlag, "", usage)
	if required {
		if err := cmd.MarkFlagRequired(JsonEntryFlag); err != nil {
			return err
		}
	}
	return nil
}

func DecorateCommandWithTTL(cmd *cobra.Command, ttl *time.Duration, required bool, usage string) error {
	cmd.Flags().DurationVar(ttl, TTLFlag, 0, usage)
	if required {
		if err := cmd.MarkFlagRequired(TTLFlag); err != nil {
			return err
		}
	}
	return nil
}

func DecorateCommandWithMaxIdle(cmd *cobra.Command, maxIdle *time.Duration, required bool, usage string) error {
	cmd.Flags().DurationVar(maxIdle, MaxIdleFlag, 0, usage)
	if required {
		if err := cmd.MarkFlagRequired(MaxIdleFlag); err != nil {
			return err
		}
	}
	return nil
}

func DecorateCommandWithDelimiter(cmd *cobra.Command, delimiter *string, required bool, usage string) error {
	cmd.Flags().StringVar(delimiter, DelimiterFlag, "\t", usage)
	if required {
		if err := cmd.MarkFlagRequired(DelimiterFlag); err != nil {
			return err
		}
	}
	return nil
}
