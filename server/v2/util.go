package serverv2

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	corectx "cosmossdk.io/core/context"
	corelog "cosmossdk.io/core/log"
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

	cmdCtx = context.WithValue(cmdCtx, corectx.LoggerContextKey, logger)
	cmdCtx = context.WithValue(cmdCtx, corectx.ViperContextKey, viper)
	cmd.SetContext(cmdCtx)

	return nil
}

func GetViperFromCmd(cmd *cobra.Command) *viper.Viper {
	value := cmd.Context().Value(corectx.ViperContextKey)
	v, ok := value.(*viper.Viper)
	if !ok {
		panic(fmt.Sprintf("incorrect viper type %T: expected *viper.Viper. Have you forgot to set the viper in the command context?", value))
	}
	return v
}

func GetLoggerFromCmd(cmd *cobra.Command) corelog.Logger {
	v := cmd.Context().Value(corectx.LoggerContextKey)
	logger, ok := v.(corelog.Logger)
	if !ok {
		panic(fmt.Sprintf("incorrect logger type %T: expected log.Logger. Have you forgot to set the logger in the command context?", v))
	}

	return logger
}
