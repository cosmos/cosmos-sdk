package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
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
	storeKey   sdk.StoreKey
	cdc        codec.BinaryMarshaler
	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper
	hooks      types.StakingHooks
	paramstore paramtypes.Subspace
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper,
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
		storeKey:   key,
		cdc:        cdc,
		authKeeper: ak,
		bankKeeper: bk,
		paramstore: ps,
		hooks:      nil,
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
	k.cdc.MustUnmarshalBinaryBare(bz, &ip)

	return ip.Int
}

// Set the last total validator power.
func (k Keeper) SetLastTotalPower(ctx sdk.Context, power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&sdk.IntProto{Int: power})
	store.Set(types.LastTotalPowerKey, bz)
}
