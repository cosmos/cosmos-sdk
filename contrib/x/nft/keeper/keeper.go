package keeper

import (
	"cosmossdk.io/core/address"
	store "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	nft2 "github.com/cosmos/cosmos-sdk/contrib/x/nft"
)

// Keeper of the nft store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	bk           nft2.BankKeeper
	ac           address.Codec
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(storeService store.KVStoreService,
	cdc codec.BinaryCodec, ak nft2.AccountKeeper, bk nft2.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft2.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		bk:           bk,
		ac:           ak.AddressCodec(),
	}
}
