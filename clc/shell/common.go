package shell

import (
	"errors"
	"fmt"
	"strings"
)

const CmdPrefix = `\`

var ErrHelp = errors.New("interactive help")

func ConvertStatement(stmt string) (string, error) {
	stmt = strings.TrimSpace(stmt)
	if strings.HasPrefix(stmt, "help") {
		return "", ErrHelp
	}
	if strings.HasPrefix(stmt, CmdPrefix) {
		// this is a shell command
		stmt = strings.TrimPrefix(stmt, CmdPrefix)
		parts := strings.Fields(stmt)
		switch parts[0] {
		case "dm":
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
			return "", fmt.Errorf("Usage: %sdm [mapping]", CmdPrefix)
		case "dm+":
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
			return "", fmt.Errorf("Usage: %sdm+ [mapping]", CmdPrefix)
		case "exit":
			return "", ErrExit
		}
		return "", fmt.Errorf("Unknown shell command: %s", stmt)
	}
	return stmt, nil
}

func InteractiveHelp() string {
	return `
Shortcut Commands:
	\dm           List mappings
	\dm  MAPPING  Display information about a mapping
	\dm+ MAPPING  Describe a mapping
	\exit         Exit the shell
	\help         Display help for CLC commands
`
}
