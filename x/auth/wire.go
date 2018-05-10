package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec for default AppAccount
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterInterface((*sdk.Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "auth/Account", nil)
}
