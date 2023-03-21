package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// Keeper manages state of all fee grants, as well as calculating approval.
// It must have a codec with all available allowances registered.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	authKeeper feegrant.AccountKeeper
}

var _ ante.FeegrantKeeper = &Keeper{}

// NewKeeper creates a fee grant Keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, ak feegrant.AccountKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		authKeeper: ak,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", feegrant.ModuleName))
}

// GrantAllowance creates a new grant
func (k Keeper) GrantAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error {
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

	store := ctx.KVStore(k.storeKey)
	key := feegrant.FeeAllowanceKey(granter, grantee)

	exp, err := feeAllowance.ExpiresAt()
	if err != nil {
		return err
	}

	// expiration shouldn't be in the past.
	if exp != nil && exp.Before(ctx.BlockTime()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "expiration is before current block time")
	}

	// if expiry is not nil, add the new key to pruning queue.
	if exp != nil {
		// `key` formed here with the prefix of `FeeAllowanceKeyPrefix` (which is `0x00`)
		// remove the 1st byte and reuse the remaining key as it is
		k.addToFeeAllowanceQueue(ctx, key[1:], exp)
	}

	grant, err := feegrant.NewGrant(granter, grantee, feeAllowance)
	if err != nil {
		return err
	}

	bz, err := k.cdc.Marshal(&grant)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			feegrant.EventTypeSetFeeGrant,
			sdk.NewAttribute(feegrant.AttributeKeyGranter, grant.Granter),
			sdk.NewAttribute(feegrant.AttributeKeyGrantee, grant.Grantee),
		),
	)

	return nil
}

// UpdateAllowance updates the existing grant.
func (k Keeper) UpdateAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error {
	store := ctx.KVStore(k.storeKey)
	key := feegrant.FeeAllowanceKey(granter, grantee)

	_, err := k.getGrant(ctx, granter, grantee)
	if err != nil {
		return err
	}

	grant, err := feegrant.NewGrant(granter, grantee, feeAllowance)
	if err != nil {
		return err
	}

	bz, err := k.cdc.Marshal(&grant)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			feegrant.EventTypeUpdateFeeGrant,
			sdk.NewAttribute(feegrant.AttributeKeyGranter, grant.Granter),
			sdk.NewAttribute(feegrant.AttributeKeyGrantee, grant.Grantee),
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
	key := feegrant.FeeAllowanceKey(granter, grantee)
	store.Delete(key)

	ctx.EventManager().EmitEvent(
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
func (k Keeper) GetAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error) {
	grant, err := k.getGrant(ctx, granter, grantee)
	if err != nil {
		return nil, err
	}

	return grant.GetGrant()
}

// getGrant returns entire grant between both accounts
func (k Keeper) getGrant(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (*feegrant.Grant, error) {
	store := ctx.KVStore(k.storeKey)
	key := feegrant.FeeAllowanceKey(granter, grantee)
	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, sdkerrors.ErrNotFound.Wrap("fee-grant not found")
	}

	var feegrant feegrant.Grant
	if err := k.cdc.Unmarshal(bz, &feegrant); err != nil {
		return nil, err
	}

	return &feegrant, nil
}

// IterateAllFeeAllowances iterates over all the grants in the store.
// Callback to get all data, returns true to stop, false to keep reading
// Calling this without pagination is very expensive and only designed for export genesis
func (k Keeper) IterateAllFeeAllowances(ctx sdk.Context, cb func(grant feegrant.Grant) bool) error {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, feegrant.FeeAllowanceKeyPrefix)
	defer iter.Close()

	stop := false
	for ; iter.Valid() && !stop; iter.Next() {
		bz := iter.Value()
		var feeGrant feegrant.Grant
		if err := k.cdc.Unmarshal(bz, &feeGrant); err != nil {
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
	return k.UpdateAllowance(ctx, granter, grantee, grant)
}

func emitUseGrantEvent(ctx sdk.Context, granter, grantee string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			feegrant.EventTypeUseFeeGrant,
			sdk.NewAttribute(feegrant.AttributeKeyGranter, granter),
			sdk.NewAttribute(feegrant.AttributeKeyGrantee, grantee),
		),
	)
}

// InitGenesis will initialize the keeper from a *previously validated* GenesisState
func (k Keeper) InitGenesis(ctx sdk.Context, data *feegrant.GenesisState) error {
	for _, f := range data.Allowances {
		granter, err := k.authKeeper.StringToBytes(f.Granter)
		if err != nil {
			return err
		}
		grantee, err := k.authKeeper.StringToBytes(f.Grantee)
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
func (k Keeper) ExportGenesis(ctx sdk.Context) (*feegrant.GenesisState, error) {
	var grants []feegrant.Grant

	err := k.IterateAllFeeAllowances(ctx, func(grant feegrant.Grant) bool {
		grants = append(grants, grant)
		return false
	})

	return &feegrant.GenesisState{
		Allowances: grants,
	}, err
}

func (k Keeper) addToFeeAllowanceQueue(ctx sdk.Context, grantKey []byte, exp *time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(feegrant.FeeAllowancePrefixQueue(exp, grantKey), []byte{})
}

// RemoveExpiredAllowances iterates grantsByExpiryQueue and deletes the expired grants.
func (k Keeper) RemoveExpiredAllowances(ctx sdk.Context) {
	exp := ctx.BlockTime()
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(feegrant.FeeAllowanceQueueKeyPrefix, storetypes.InclusiveEndBytes(feegrant.AllowanceByExpTimeKey(&exp)))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())

		granter, grantee := feegrant.ParseAddressesFromFeeAllowanceQueueKey(iterator.Key())
		store.Delete(feegrant.FeeAllowanceKey(granter, grantee))
	}
}
