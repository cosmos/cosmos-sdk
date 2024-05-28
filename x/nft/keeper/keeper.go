package keeper

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/nft"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper of the nft store
type Keeper struct {
	appmodule.Environment

	cdc codec.BinaryCodec
	bk  nft.BankKeeper
	ac  address.Codec
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(env appmodule.Environment,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		Environment: env,
		cdc:         cdc,
		bk:          bk,
		ac:          ak.AddressCodec(),
	}
}
