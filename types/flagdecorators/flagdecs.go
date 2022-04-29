package flagdecorators

import (
	"github.com/spf13/cobra"
)

func DecorateCommandWithAllFlag(cmd *cobra.Command, all *bool, required bool, usage string) {
	cmd.Flags().BoolVar(all, "all", false, usage)
	if required {
		cmd.MarkFlagRequired("all")
	}
}
