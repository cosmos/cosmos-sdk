package server

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/version"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tmlibs/cli"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	"github.com/tendermint/tmlibs/log"
)

type Context struct {
	Config *cfg.Config
	Logger log.Logger
}

func NewDefaultContext() *Context {
	return NewContext(
		cfg.DefaultConfig(),
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	)
}

func NewContext(config *cfg.Config, logger log.Logger) *Context {
	return &Context{config, logger}
}

//--------------------------------------------------------------------

// PersistentPreRunEFn returns a PersistentPreRunE function for cobra
// that initailizes the passed in context with a properly configured
// logger and config objecy
func PersistentPreRunEFn(context *Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == version.VersionCmd.Name() {
			return nil
		}
		config, err := tcmd.ParseConfig()
		if err != nil {
			return err
		}
		logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		context.Config = config
		context.Logger = logger
		return nil
	}
}

func AddCommands(
	rootCmd *cobra.Command,
	appState GenAppState, appCreator AppCreator,
	context *Context) {

	rootCmd.PersistentFlags().String("log_level", context.Config.LogLevel, "Log level")

	rootCmd.AddCommand(
		InitCmd(appState, context),
		StartCmd(appCreator, context),
		UnsafeResetAllCmd(context),
		ShowNodeIDCmd(context),
		ShowValidatorCmd(context),
		version.VersionCmd,
	)
}
