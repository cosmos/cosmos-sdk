package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	store2 "github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// Keeper of the mint store
type Keeper struct {
	cdc              codec.BinaryCodec
	storeKey         storetypes.StoreKey
	stakingKeeper    types.StakingKeeper
	bankKeeper       types.BankKeeper
	feeCollectorName string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	sk types.StakingKeeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("the x/%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		stakingKeeper:    sk,
		bankKeeper:       bk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
	}
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) decodeMinter(bz []byte) (types.Minter, error) {
	var minter types.Minter
	if bz == nil {
		panic("stored minter should not have been nil")
	}
	k.cdc.MustUnmarshal(bz, &minter)
	return minter, nil
}

func (k Keeper) getStore(ctx sdk.Context) store2.StoreAPI {
	return store2.NewStoreAPI(ctx.KVStore(k.storeKey))
}

// get the minter
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	minter, _ = store2.GetAndDecode(store, k.decodeMinter, types.MinterKey)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&minter)
	store.Set(types.MinterKey, bz)
}

// SetParams sets the x/mint module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)

	return nil
}

func (k Keeper) decodeParams(bz []byte) (types.Params, error) {
	var params types.Params
	if bz == nil {
		return params, nil
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params, nil
}

// GetParams returns the current x/mint module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	params, _ := store2.GetAndDecode(store, k.decodeParams, types.ParamsKey)
	return params
}

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx sdk.Context) math.Int {
	return k.stakingKeeper.StakingTokenSupply(ctx)
}

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx sdk.Context) math.LegacyDec {
	return k.stakingKeeper.BondedRatio(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx sdk.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}
