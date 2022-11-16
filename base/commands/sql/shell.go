//go:build base || sql

package sql

import (
	"errors"
	"fmt"
	"strings"
)

func convertStatement(stmt string) (string, error) {
	stmt = strings.TrimSpace(stmt)
	if strings.HasPrefix(stmt, "help") {
		return "", errors.New(interactiveHelp())
	}
	if strings.HasPrefix(stmt, "\\") {
		// this is a shell command
		parts := strings.Fields(stmt)
		switch parts[0] {
		case "\\dm":
			if len(parts) == 1 {
				return "show mappings;", nil
			}
			if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				return fmt.Sprintf(`
					SELECT * FROM information_schema.mappings
					WHERE table_name = '%s';
				`, mn), nil
			}
			return "", fmt.Errorf("Usage: \\dm [mapping]")
		case "\\dm+":
			if len(parts) == 1 {
				return "show mappings;", nil
			}
			if len(parts) == 2 {
				// escape single quote
				mn := strings.Replace(parts[1], "'", "''", -1)
				return fmt.Sprintf(`
					SELECT * FROM information_schema.columns
					WHERE table_name = '%s';
				`, mn), nil
			}
			return "", fmt.Errorf("Usage: \\dm+ [mapping]")
		}
		return "", fmt.Errorf("Unknown shell command: %s", stmt)
	}
	return stmt, nil
}

func interactiveHelp() string {
	return `
Commands:
	\dm           list mappings
	\dm  MAPPING  display info about a mapping
	\dm+ MAPPING  describe a mapping
`
}
