package cobra

import (
	"io"

	"github.com/spf13/cobra"
)

func InitCommandForCustomInvocation(rootCmd *cobra.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer, args []string) {
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)
}
