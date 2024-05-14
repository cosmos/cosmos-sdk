package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the mint store
type Keeper struct {
	appmodule.Environment

	cdc              codec.BinaryCodec
	bankKeeper       types.BankKeeper
	logger           log.Logger
	feeCollectorName string
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	Schema             collections.Schema
	Params             collections.Item[types.Params]
	Minter             collections.Item[types.Minter]
	LastReductionEpoch collections.Item[int64]
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	env appmodule.Environment,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("the x/%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(env.KVStoreService)
	k := Keeper{
		Environment:        env,
		cdc:                cdc,
		bankKeeper:         bk,
		logger:             env.Logger,
		feeCollectorName:   feeCollectorName,
		authority:          authority,
		Params:             collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Minter:             collections.NewItem(sb, types.MinterKey, "minter", codec.CollValue[types.Minter](cdc)),
		LastReductionEpoch: collections.NewItem(sb, types.LastReductionEpochKey, "last_reduction_epoch", collections.Int64Value),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in epochs hooks for minting.
func (k Keeper) MintCoins(ctx context.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in epochs hooks for minting.
func (k Keeper) AddCollectedFees(ctx context.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}

func (k Keeper) setLastReductionEpochNum(ctx context.Context, epochNum int64) error {
	return k.LastReductionEpoch.Set(ctx, epochNum)
}

func (k Keeper) getLastReductionEpochNum(ctx context.Context) (int64, error) {
	return k.LastReductionEpoch.Get(ctx)
}
