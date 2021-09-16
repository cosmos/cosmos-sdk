package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Implements ValidatorSet interface
var _ types.ValidatorSet = Keeper{}

// Implements DelegationSet interface
var _ types.DelegationSet = Keeper{}

// keeper of the staking store
type Keeper struct {
	storeKey     sdk.StoreKey
	transientKey sdk.StoreKey
	cdc          codec.BinaryCodec
	authKeeper   types.AccountKeeper
	bankKeeper   types.BankKeeper
	hooks        types.StakingHooks
	paramstore   paramtypes.Subspace
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, key, tKey sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper,
	ps paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	// ensure bonded and not bonded module accounts are set
	if addr := ak.GetModuleAddress(types.BondedPoolName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.BondedPoolName))
	}

	if addr := ak.GetModuleAddress(types.NotBondedPoolName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.NotBondedPoolName))
	}

	return Keeper{
		storeKey:     key,
		transientKey: tKey,
		cdc:          cdc,
		authKeeper:   ak,
		bankKeeper:   bk,
		paramstore:   ps,
		hooks:        nil,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Set the validator hooks
func (k *Keeper) SetHooks(sh types.StakingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = sh

	return k
}

// Load the last total validator power.
func (k Keeper) GetLastTotalPower(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastTotalPowerKey)

	if bz == nil {
		return sdk.ZeroInt()
	}

	ip := sdk.IntProto{}
	k.cdc.MustUnmarshal(bz, &ip)

	return ip.Int
}

// Set the last total validator power.
func (k Keeper) SetLastTotalPower(ctx sdk.Context, power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.IntProto{Int: power})
	store.Set(types.LastTotalPowerKey, bz)
}

// GetValidatorUpdate returns the ABCI validator power update for the current block
// by the consensus address.
func (k Keeper) GetValidatorUpdate(ctx sdk.Context, consAddr sdk.ConsAddress) (abci.ValidatorUpdate, bool) {
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.ValidatorUpdatesKey)
	bz := store.Get(consAddr.Bytes())
	if len(bz) == 0 {
		return abci.ValidatorUpdate{}, false
	}

	var valUpdate abci.ValidatorUpdate
	k.cdc.MustUnmarshal(bz, &valUpdate)
	return valUpdate, true
}

// HasValidatorUpdate returns true if there is a power update for the given validator
// within the last block.
func (k Keeper) HasValidatorUpdate(ctx sdk.Context, consAddr sdk.ConsAddress) bool {
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.ValidatorUpdatesKey)
	return store.Has(consAddr.Bytes())
}

// SetValidatorUpdate sets the ABCI validator power update for the current block
// by the consensus address.
func (k Keeper) SetValidatorUpdate(ctx sdk.Context, consAddr sdk.ConsAddress, valUpdate abci.ValidatorUpdate) {
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.ValidatorUpdatesKey)
	bz := k.cdc.MustMarshal(&valUpdate)
	store.Set(consAddr.Bytes(), bz)
}

// GetValidatorUpdates returns all the ABCI validator power updates within the current block.
func (k Keeper) GetValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	store := ctx.TransientStore(k.transientKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorUpdatesKey)

	var valsetUpdates []abci.ValidatorUpdate
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var valUpdate abci.ValidatorUpdate
		k.cdc.MustUnmarshal(iterator.Value(), &valUpdate)
		valsetUpdates = append(valsetUpdates, valUpdate)
	}

	return valsetUpdates
}
