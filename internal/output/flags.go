package output

import (
	"fmt"

	"github.com/spf13/cobra"
)

func TypeStringFor(cmd *cobra.Command) (Type, error) {
	f := cmd.Flag("output-type")
	if f == nil {
		return TypeDelimited, nil
	}
	switch f.Value.String() {
	case TypeStringDefault:
		return TypeDelimited, nil
	case TypeStringJSON:
		return TypeJSON, nil
	case TypeStringCSV:
		return TypeCSV, nil
	case TypeStringPretty:
		return TypeTable, nil
	}
	return TypeDelimited, fmt.Errorf("invalid output type: %s", f.Value.String())
}
