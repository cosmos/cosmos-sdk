package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// Keeper manages state of all fee grants, as well as calculating approval.
// It must have a codec with all available allowances registered.
type Keeper struct {
	cdc               codec.BinaryCodec
	storeService      store.KVStoreService
	authKeeper        feegrant.AccountKeeper
	Schema            collections.Schema
	FeeAllowance      collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], feegrant.Grant]
	FeeAllowanceQueue collections.Map[collections.Triple[time.Time, sdk.AccAddress, sdk.AccAddress], bool]
}

var _ ante.FeegrantKeeper = &Keeper{}

// NewKeeper creates a feegrant Keeper
func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, ak feegrant.AccountKeeper) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authKeeper:   ak,
		FeeAllowance: collections.NewMap(
			sb,
			feegrant.FeeAllowanceKeyPrefix,
			"allowances",
			collections.PairKeyCodec(sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[feegrant.Grant](cdc),
		),
		FeeAllowanceQueue: collections.NewMap(
			sb,
			feegrant.FeeAllowanceQueueKeyPrefix,
			"allowances_queue",
			collections.TripleKeyCodec(sdk.TimeKey, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collections.BoolValue,
		),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", feegrant.ModuleName))
}

// GrantAllowance creates a new grant
func (k Keeper) GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error {
	// Checking for duplicate entry
	if f, _ := k.GetAllowance(ctx, granter, grantee); f != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "fee allowance already exists")
	}

	// create the account if it is not in account state
	granteeAcc := k.authKeeper.GetAccount(ctx, grantee)
	if granteeAcc == nil {
		granteeAcc = k.authKeeper.NewAccountWithAddress(ctx, grantee)
		k.authKeeper.SetAccount(ctx, granteeAcc)
	}

	exp, err := feeAllowance.ExpiresAt()
	if err != nil {
		return err
	}

	// expiration shouldn't be in the past.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if exp != nil && exp.Before(sdkCtx.BlockTime()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "expiration is before current block time")
	}

	// if expiry is not nil, add the new key to pruning queue.
	if exp != nil {
		err = k.FeeAllowanceQueue.Set(ctx, collections.Join3(*exp, grantee, granter), true)
		if err != nil {
			return err
		}
	}

	grant, err := feegrant.NewGrant(granter, grantee, feeAllowance)
	if err != nil {
		return err
	}

	if err := k.FeeAllowance.Set(ctx, collections.Join(grantee, granter), grant); err != nil {
		return err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			feegrant.EventTypeSetFeeGrant,
			sdk.NewAttribute(feegrant.AttributeKeyGranter, grant.Granter),
			sdk.NewAttribute(feegrant.AttributeKeyGrantee, grant.Grantee),
		),
	)

	return nil
}

// UpdateAllowance updates the existing grant.
func (k Keeper) UpdateAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error {
	_, err := k.getGrant(ctx, granter, grantee)
	if err != nil {
		return err
	}

	grant, err := feegrant.NewGrant(granter, grantee, feeAllowance)
	if err != nil {
		return err
	}

	if err := k.FeeAllowance.Set(ctx, collections.Join(grantee, granter), grant); err != nil {
		return err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
		sdk.NewEvent(
			feegrant.EventTypeUpdateFeeGrant,
			sdk.NewAttribute(feegrant.AttributeKeyGranter, grant.Granter),
			sdk.NewAttribute(feegrant.AttributeKeyGrantee, grant.Grantee),
		),
	)

	return nil
}

// revokeAllowance removes an existing grant
func (k Keeper) revokeAllowance(ctx context.Context, granter, grantee sdk.AccAddress) error {
	grant, err := k.GetAllowance(ctx, granter, grantee)
	if err != nil {
		return err
	}

	if err := k.FeeAllowance.Remove(ctx, collections.Join(grantee, granter)); err != nil {
		return err
	}

	exp, err := grant.ExpiresAt()
	if err != nil {
		return err
	}

	if exp != nil {
		if err := k.FeeAllowanceQueue.Remove(ctx, collections.Join3(*exp, grantee, granter)); err != nil {
			return err
		}
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
		sdk.NewEvent(
			feegrant.EventTypeRevokeFeeGrant,
			sdk.NewAttribute(feegrant.AttributeKeyGranter, granter.String()),
			sdk.NewAttribute(feegrant.AttributeKeyGrantee, grantee.String()),
		),
	)
	return nil
}

// GetAllowance returns the allowance between the granter and grantee.
// If there is none, it returns nil, nil.
// Returns an error on parsing issues
func (k Keeper) GetAllowance(ctx context.Context, granter, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error) {
	grant, err := k.FeeAllowance.Get(ctx, collections.Join(grantee, granter))
	if err != nil {
		return nil, err
	}

	return grant.GetGrant()
}

// getGrant returns entire grant between both accounts
func (k Keeper) getGrant(ctx context.Context, granter, grantee sdk.AccAddress) (*feegrant.Grant, error) {
	feegrant, err := k.FeeAllowance.Get(ctx, collections.Join(grantee, granter))
	if err != nil {
		return nil, err
	}

	return &feegrant, nil
}

// IterateAllFeeAllowances iterates over all the grants in the store.
// Callback to get all data, returns true to stop, false to keep reading
// Calling this without pagination is very expensive and only designed for export genesis
func (k Keeper) IterateAllFeeAllowances(ctx context.Context, cb func(grant feegrant.Grant) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iter := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), feegrant.FeeAllowanceKeyPrefix)
	defer iter.Close()

	err := k.FeeAllowance.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], grant feegrant.Grant) (stop bool, err error) {
		return cb(grant), nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UseGrantedFees will try to pay the given fee from the granter's account as requested by the grantee
func (k Keeper) UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
	f, err := k.getGrant(ctx, granter, grantee)
	if err != nil {
		return err
	}

	grant, err := f.GetGrant()
	if err != nil {
		return err
	}

	remove, err := grant.Accept(ctx, fee, msgs)
	if remove {
		// Ignoring the `revokeFeeAllowance` error, because the user has enough grants to perform this transaction.
		_ = k.revokeAllowance(ctx, granter, grantee)
		if err != nil {
			return err
		}
		emitUseGrantEvent(ctx, granter.String(), grantee.String())

		return nil
	}
	if err != nil {
		return err
	}
	emitUseGrantEvent(ctx, granter.String(), grantee.String())

	// if fee allowance is accepted, store the updated state of the allowance
	return k.UpdateAllowance(ctx, granter, grantee, grant)
}

func emitUseGrantEvent(ctx context.Context, granter, grantee string) {
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
		sdk.NewEvent(
			feegrant.EventTypeUseFeeGrant,
			sdk.NewAttribute(feegrant.AttributeKeyGranter, granter),
			sdk.NewAttribute(feegrant.AttributeKeyGrantee, grantee),
		),
	)
}

// InitGenesis will initialize the keeper from a *previously validated* GenesisState
func (k Keeper) InitGenesis(ctx context.Context, data *feegrant.GenesisState) error {
	for _, f := range data.Allowances {
		granter, err := k.authKeeper.AddressCodec().StringToBytes(f.Granter)
		if err != nil {
			return err
		}
		grantee, err := k.authKeeper.AddressCodec().StringToBytes(f.Grantee)
		if err != nil {
			return err
		}

		grant, err := f.GetGrant()
		if err != nil {
			return err
		}

		err = k.GrantAllowance(ctx, granter, grantee, grant)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis will dump the contents of the keeper into a serializable GenesisState.
func (k Keeper) ExportGenesis(ctx context.Context) (*feegrant.GenesisState, error) {
	var grants []feegrant.Grant

	err := k.IterateAllFeeAllowances(ctx, func(grant feegrant.Grant) bool {
		grants = append(grants, grant)
		return false
	})

	return &feegrant.GenesisState{
		Allowances: grants,
	}, err
}

// RemoveExpiredAllowances iterates grantsByExpiryQueue and deletes the expired grants.
func (k Keeper) RemoveExpiredAllowances(ctx context.Context) error {
	exp := sdk.UnwrapSDKContext(ctx).BlockTime()
	rng := collections.NewPrefixUntilTripleRange[time.Time, sdk.AccAddress, sdk.AccAddress](exp)

	err := k.FeeAllowanceQueue.Walk(ctx, rng, func(key collections.Triple[time.Time, sdk.AccAddress, sdk.AccAddress], value bool) (stop bool, err error) {
		grantee, granter := key.K2(), key.K3()

		if err := k.FeeAllowance.Remove(ctx, collections.Join(grantee, granter)); err != nil {
			return true, err
		}

		if err := k.FeeAllowanceQueue.Remove(ctx, key); err != nil {
			return true, err
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}
