package context

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
)

const ServerContextKey ContextKey = "server.context"

type ServerContext interface {
	GetLogger() log.Logger
	GetViper() *viper.Viper
	GetConfig() *cmtcfg.Config
	SetRoot(string)
}

func GetServerContextFromCmd(cmd *cobra.Command) ServerContext {
	if v := cmd.Context().Value(ServerContextKey); v != nil {
		serverCtxPtr := v.(ServerContext)
		return serverCtxPtr
	}
	return nil
}
