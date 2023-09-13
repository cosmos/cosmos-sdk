package internal

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/hubl/internal/flags"
)

func VersionCmd() *cobra.Command {
	var long bool

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display hubl version",
		RunE: func(cmd *cobra.Command, args []string) error {
			version, ok := debug.ReadBuildInfo()
			if !ok {
				return fmt.Errorf("failed to get hubl version")
			}

			cmd.Printf("hubl version: %s\n", strings.TrimSpace(version.Main.Version))

			if long {
				for _, dep := range version.Deps {
					if dep.Path == "cosmossdk.io/client/v2" {
						cmd.Printf("client/v2 version: %s\n", strings.TrimSpace(dep.Version))
					}
				}
			}

			return nil
		},
	}

	versionCmd.Flags().BoolVar(&long, flags.FlagLong, false, "display long version information")

	return versionCmd
}
