package helpers

import "github.com/spf13/cobra"

func DefaultCmd() *cobra.Command {
	cmd := &cobra.Command{}
	pflags := cmd.PersistentFlags()
	pflags.String(FlagLogLevel, "info", "The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>')")
	pflags.String(FlagLogFormat, "plain", "The logging format (json|plain)")
	pflags.Bool(FlagLogNoColor, false, "Disable colored logs")
	pflags.StringP(FlagHome, "", defaultHome, "directory for config and data")
}
