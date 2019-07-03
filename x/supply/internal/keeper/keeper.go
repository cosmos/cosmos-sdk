package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// Keeper of the supply store
type Keeper struct {
	cdc       *codec.Codec
	storeKey  sdk.StoreKey
	ak        types.AccountKeeper
	bk        types.BankKeeper
	permAddrs map[string]permAddr
}

type permAddr struct {
	permission string // basic/minter/burner
	address    sdk.AccAddress
}

// NewpermAddr creates a new permAddr object
func newPermAddr(permission, name string) permAddr {
	return permAddr{
		permission: permission,
		address:    types.NewModuleAddress(name),
	}
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper,
	codespace sdk.CodespaceType, basicModuleAccs, minterModuleAccs, burnerModuleAccs []string) Keeper {

	// set the addresses
	permAddrs := make(map[string]permAddr)
	for _, name := range basicModuleAccs {
		permAddrs[name] = newPermAddr(types.Basic, name)
	}
	for _, name := range minterModuleAccs {
		permAddrs[name] = newPermAddr(types.Minter, name)
	}
	for _, name := range burnerModuleAccs {
		permAddrs[name] = newPermAddr(types.Burner, name)
	}

	return Keeper{
		cdc:       cdc,
		storeKey:  key,
		ak:        ak,
		bk:        bk,
		permAddrs: permAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetSupply retrieves the Supply from store
func (k Keeper) GetSupply(ctx sdk.Context) (supply types.Supply) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(SupplyKey)
	if b == nil {
		panic("stored supply should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &supply)
	return
}

// SetSupply sets the Supply to store
func (k Keeper) SetSupply(ctx sdk.Context, supply types.Supply) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(supply)
	store.Set(SupplyKey, b)
}
