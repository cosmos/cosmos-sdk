package stateviewer

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display stateviewer version",
		RunE: func(cmd *cobra.Command, args []string) error {
			version, ok := debug.ReadBuildInfo()
			if !ok {
				return fmt.Errorf("failed to get stateviewer version")
			}

			cmd.Printf("stateviewer version: %s\n", strings.TrimSpace(version.Main.Version))
			return nil
		},
	}
}
