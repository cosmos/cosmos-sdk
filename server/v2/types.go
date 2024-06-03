package serverv2

import (
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type Application[T transaction.Tx] interface {
	GetAppManager() *appmanager.AppManager[T]
	GetConsensusAuthority() string
	InterfaceRegistry() codectypes.InterfaceRegistry
	// GetStore() any
}
