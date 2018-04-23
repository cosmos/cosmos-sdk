package server

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/wire"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tmlibs/cli"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	"github.com/tendermint/tmlibs/log"
)

// server context
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

//___________________________________________________________________________________

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

// add server commands
func AddCommands(
	ctx *Context, cdc *wire.Codec,
	rootCmd *cobra.Command,
	appState GenAppParams, appCreator AppCreator) {

	rootCmd.PersistentFlags().String("log_level", ctx.Config.LogLevel, "Log level")

	rootCmd.AddCommand(
		InitCmd(ctx, cdc, appState, nil),
		StartCmd(ctx, appCreator),
		UnsafeResetAllCmd(ctx),
		ShowNodeIDCmd(ctx),
		ShowValidatorCmd(ctx),
		version.VersionCmd,
	)
}

//___________________________________________________________________________________

// append a new json field to existing json message
func AppendJSON(cdc *wire.Codec, baseJSON []byte, key string, value json.RawMessage) (appended []byte, err error) {
	var jsonMap map[string]json.RawMessage
	err = cdc.UnmarshalJSON(baseJSON, &jsonMap)
	if err != nil {
		return nil, err
	}
	jsonMap[key] = value
	bz, err := wire.MarshalJSONIndent(cdc, jsonMap)
	return json.RawMessage(bz), err
}
