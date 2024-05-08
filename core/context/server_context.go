package context

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
)

type ContextKey string

const ServerContextKey ContextKey = "server.context"

type ServerContext interface {
	GetLogger() log.Logger
	GetViper() *viper.Viper
	GetConfig() CometConfig
	SetRoot(rootDir string)
}

type BaseConfig interface {
	DBDir() string
	GenesisFile() string
	NodeKeyFile() string
	PrivValidatorKeyFile() string
	PrivValidatorStateFile() string
}

type CometConfig interface {
	BaseConfig
	CheckDeprecated() []string
	SetRoot(root string) CometConfig
	ValidateBasic() error
}

func GetServerContextFromCmd(cmd *cobra.Command) ServerContext {
	if v := cmd.Context().Value(ServerContextKey); v != nil {
		serverCtxPtr := v.(ServerContext)
		return serverCtxPtr
	}
	return nil
}
