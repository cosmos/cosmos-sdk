// Deprecated: This package is deprecated and will be removed in the next major release. The `x/nft` module will be moved to a separate repo `github.com/cosmos/cosmos-sdk-legacy`.
package keeper

import (
	"cosmossdk.io/core/address"
	store "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/nft" //nolint:staticcheck // deprecated and to be removed
)

// Keeper of the nft store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	bk           nft.BankKeeper
	ac           address.Codec
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(storeService store.KVStoreService,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		bk:           bk,
		ac:           ak.AddressCodec(),
	}
}
