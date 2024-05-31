package serverv2

import (
	"context"
	"os"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	corectx "cosmossdk.io/core/context"
	"cosmossdk.io/log"
)

// SetCmdServerContext sets a command's Context value to the provided argument.
// If the context has not been set, set the given context as the default.
func SetCmdServerContext(cmd *cobra.Command, viper *viper.Viper, logger log.Logger) error {
	var cmdCtx context.Context

	if cmd.Context() == nil {
		cmdCtx = context.Background()
	} else {
		cmdCtx = cmd.Context()
	}

	cmd.SetContext(context.WithValue(cmdCtx, corectx.LoggerContextKey{}, logger))
	cmd.SetContext(context.WithValue(cmdCtx, corectx.ViperContextKey{}, viper))

	return nil
}

func GetViperFromCmd(cmd *cobra.Command) *viper.Viper {
	value := cmd.Context().Value(corectx.ViperContextKey{})
	v, ok := value.(*viper.Viper)
	if !ok {
		return viper.New()
	}
	return v
}

func GetConfigFromViper(v *viper.Viper) *cmtcfg.Config {
	conf := cmtcfg.DefaultConfig()
	err := v.Unmarshal(conf)
	rootDir := v.GetString("home")
	if err != nil {
		return cmtcfg.DefaultConfig().SetRoot(rootDir)
	}
	return conf.SetRoot(rootDir)
}

func GetLoggerFromCmd(cmd *cobra.Command) log.Logger {
	v := cmd.Context().Value(corectx.LoggerContextKey{})
	logger, ok := v.(log.Logger)
	if !ok {
		return log.NewLogger(os.Stdout)
	}
	return logger
}
