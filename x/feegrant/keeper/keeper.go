package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// Keeper manages state of all fee grants, as well as calculating approval.
// It must have a codec with all available allowances registered.
type Keeper struct {
	cdc        codec.BinaryMarshaler
	storeKey   sdk.StoreKey
	authKeeper types.AccountKeeper
}

var _ ante.FeegrantKeeper = &Keeper{}

// NewKeeper creates a fee grant Keeper
func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, ak types.AccountKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		authKeeper: ak,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GrantFeeAllowance creates a new grant
func (k Keeper) GrantFeeAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress, feeAllowance types.FeeAllowanceI) error {

	// create the account if it is not in account state
	granteeAcc := k.authKeeper.GetAccount(ctx, grantee)
	if granteeAcc == nil {
		granteeAcc = k.authKeeper.NewAccountWithAddress(ctx, grantee)
		k.authKeeper.SetAccount(ctx, granteeAcc)
	}

	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	grant, err := types.NewFeeAllowanceGrant(granter, grantee, feeAllowance)
	if err != nil {
		return err
	}

	bz, err := k.cdc.MarshalBinaryBare(&grant)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetFeeGrant,
			sdk.NewAttribute(types.AttributeKeyGranter, grant.Granter),
			sdk.NewAttribute(types.AttributeKeyGrantee, grant.Grantee),
		),
	)

	return nil
}

// RevokeFeeAllowance removes an existing grant
func (k Keeper) RevokeFeeAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	_, err := k.getFeeGrant(ctx, granter, grantee)
	if err != nil {
		return err
	}

	store.Delete(key)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRevokeFeeGrant,
			sdk.NewAttribute(types.AttributeKeyGranter, granter.String()),
			sdk.NewAttribute(types.AttributeKeyGrantee, grantee.String()),
		),
	)
	return nil
}

// GetFeeAllowance returns the allowance between the granter and grantee.
// If there is none, it returns nil, nil.
// Returns an error on parsing issues
func (k Keeper) GetFeeAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) (types.FeeAllowanceI, error) {
	grant, err := k.getFeeGrant(ctx, granter, grantee)
	if err != nil {
		return nil, err
	}

	return grant.GetFeeGrant()
}

// getFeeGrant returns entire grant between both accounts
func (k Keeper) getFeeGrant(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (*types.FeeAllowanceGrant, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "fee-grant not found")
	}

	var feegrant types.FeeAllowanceGrant
	if err := k.cdc.UnmarshalBinaryBare(bz, &feegrant); err != nil {
		return nil, err
	}

	return &feegrant, nil
}

// IterateAllFeeAllowances iterates over all the grants in the store.
// Callback to get all data, returns true to stop, false to keep reading
// Calling this without pagination is very expensive and only designed for export genesis
func (k Keeper) IterateAllFeeAllowances(ctx sdk.Context, cb func(types.FeeAllowanceGrant) bool) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FeeAllowanceKeyPrefix)
	defer iter.Close()

	stop := false
	for ; iter.Valid() && !stop; iter.Next() {
		bz := iter.Value()
		var feeGrant types.FeeAllowanceGrant
		if err := k.cdc.UnmarshalBinaryBare(bz, &feeGrant); err != nil {
			return err
		}

		stop = cb(feeGrant)
	}

	return nil
}

// UseGrantedFees will try to pay the given fee from the granter's account as requested by the grantee
func (k Keeper) UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
	f, err := k.getFeeGrant(ctx, granter, grantee)
	if err != nil {
		return err
	}

	grant, err := f.GetFeeGrant()
	if err != nil {
		return err
	}

	remove, err := grant.Accept(ctx, fee, msgs)

	if remove {
		k.RevokeFeeAllowance(ctx, granter, grantee)
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
	return k.GrantFeeAllowance(ctx, granter, grantee, grant)
}

func emitUseGrantEvent(ctx sdk.Context, granter, grantee string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUseFeeGrant,
			sdk.NewAttribute(types.AttributeKeyGranter, granter),
			sdk.NewAttribute(types.AttributeKeyGrantee, grantee),
		),
	)
}
