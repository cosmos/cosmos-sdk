package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// Keeper of the nft store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
	bk       nft.BankKeeper
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(key sdk.StoreKey,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		cdc:      cdc,
		storeKey: key,
		bk:       bk,
	}
}
