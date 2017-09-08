package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tmlibs/cli"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	"github.com/tendermint/tmlibs/log"
)

//nolint
const (
	defaultLogLevel = "error"
	FlagLogLevel    = "log_level"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
)

// preRunSetup should be set as PersistentPreRunE on the root command to
// properly handle the logging and the tracer
func preRunSetup(cmd *cobra.Command, args []string) (err error) {
	level := viper.GetString(FlagLogLevel)
	logger, err = tmflags.ParseLogLevel(level, logger, defaultLogLevel)
	if err != nil {
		return err
	}
	if viper.GetBool(cli.TraceFlag) {
		logger = log.NewTracingLogger(logger)
	}
	return nil
}

// SetUpRoot - initialize the root command
func SetUpRoot(cmd *cobra.Command) {
	cmd.PersistentPreRunE = preRunSetup
	cmd.PersistentFlags().String(FlagLogLevel, defaultLogLevel, "Log level")
}
