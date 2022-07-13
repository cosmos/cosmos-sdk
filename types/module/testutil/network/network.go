package network

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

type Validator interface {
	GetCtx() *server.Context
	GetAppConfig() *srvconfig.Config
}

// AppConstructor defines a function which accepts a network configuration and
// creates an ABCI Application to provide to Tendermint.
type AppConstructor = func(val Validator) servertypes.Application
type GenesisState map[string]json.RawMessage
