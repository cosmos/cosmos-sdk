package context

import (
	"fmt"

	"cosmossdk.io/log"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const ServerContextKey ContextKey = "server-ctx"

type ServerContext interface{
	GetLogger() log.Logger
	GetViper() *viper.Viper
	GetConfig() *cmtcfg.Config
}

func GetServerContextFromCmd(cmd *cobra.Command) ServerContext {
	if v := cmd.Context().Value(ServerContextKey); v != nil {
		fmt.Println("serverCtxPtr", v)
		serverCtxPtr := v.(ServerContext)
		return serverCtxPtr
	}
	return nil
}