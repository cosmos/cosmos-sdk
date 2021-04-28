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

// GrantAllowance creates a new grant
func (k Keeper) GrantAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress, feeAllowance types.FeeAllowanceI) error {

	// create the account if it is not in account state
	granteeAcc := k.authKeeper.GetAccount(ctx, grantee)
	if granteeAcc == nil {
		granteeAcc = k.authKeeper.NewAccountWithAddress(ctx, grantee)
		k.authKeeper.SetAccount(ctx, granteeAcc)
	}

	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	grant, err := types.NewGrant(granter, grantee, feeAllowance)
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

// revokeAllowance removes an existing grant
func (k Keeper) revokeAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) error {
	_, err := k.getGrant(ctx, granter, grantee)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
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

// GetAllowance returns the allowance between the granter and grantee.
// If there is none, it returns nil, nil.
// Returns an error on parsing issues
func (k Keeper) GetAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) (types.FeeAllowanceI, error) {
	grant, err := k.getGrant(ctx, granter, grantee)
	if err != nil {
		return nil, err
	}

	return grant.GetGrant()
}

// getGrant returns entire grant between both accounts
func (k Keeper) getGrant(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (*types.Grant, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "fee-grant not found")
	}

	var feegrant types.Grant
	if err := k.cdc.UnmarshalBinaryBare(bz, &feegrant); err != nil {
		return nil, err
	}

	return &feegrant, nil
}

// IterateAllFeeAllowances iterates over all the grants in the store.
// Callback to get all data, returns true to stop, false to keep reading
// Calling this without pagination is very expensive and only designed for export genesis
func (k Keeper) IterateAllFeeAllowances(ctx sdk.Context, cb func(grant types.Grant) bool) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FeeAllowanceKeyPrefix)
	defer iter.Close()

	stop := false
	for ; iter.Valid() && !stop; iter.Next() {
		bz := iter.Value()
		var feeGrant types.Grant
		if err := k.cdc.UnmarshalBinaryBare(bz, &feeGrant); err != nil {
			return err
		}

		stop = cb(feeGrant)
	}

	return nil
}

// UseGrantedFees will try to pay the given fee from the granter's account as requested by the grantee
func (k Keeper) UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
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
		k.revokeAllowance(ctx, granter, grantee)
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
	return k.GrantAllowance(ctx, granter, grantee, grant)
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