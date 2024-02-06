package keeper

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/x/nft"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper of the nft store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	bk           nft.BankKeeper
	ac           address.Codec
	eventService event.Service
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(env appmodule.Environment,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {

	storeService := env.KVStoreService
	
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		eventService: env.EventService,
		bk:           bk,
		ac:           ak.AddressCodec(),
	}
}
