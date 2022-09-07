package testutil

import (
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// TODO rename
type Validator interface {
	GetCtx() *server.Context
	GetAppConfig() *srvconfig.Config
}
