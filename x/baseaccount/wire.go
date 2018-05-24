package baseaccount

import (
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "baseaccount/BaseAccount", nil)
	wire.RegisterCrypto(cdc)
	cdc.RegisterConcrete(MsgChangeKey{}, "baseaccount/changekey", nil)
}
