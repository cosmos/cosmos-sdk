package keeper

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	v1 "cosmossdk.io/x/accounts/extensions/feegrant/v1"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// Keeper manages state of all fee grants, as well as calculating approval.
// It must have a codec with all available allowances registered.
type Keeper struct {
	appmodule.Environment

	cdc     codec.BinaryCodec
	addrCdc address.Codec
	Schema  collections.Schema
	// FeeAllowance key: grantee+granter | value: Grant
	FeeAllowance collections.Map[collections.Pair[sdk.AccAddress, sdk.AccAddress], feegrant.Grant]
	// FeeAllowanceQueue key: expiration time+grantee+granter | value: bool
	FeeAllowanceQueue collections.Map[collections.Triple[time.Time, sdk.AccAddress, sdk.AccAddress], bool]
	accountKeeper     feegrant.AccountsKeeper
}

var _ ante.FeegrantKeeper = &Keeper{}

// NewKeeper creates a feegrant Keeper
func NewKeeper(
	env appmodule.Environment,
	cdc codec.BinaryCodec,
	addrCdc address.Codec,
	accountKeeper feegrant.AccountsKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)

	return Keeper{
		Environment:   env,
		accountKeeper: accountKeeper,
		cdc:           cdc,
		addrCdc:       addrCdc,
		FeeAllowance: collections.NewMap(
			sb,
			feegrant.FeeAllowanceKeyPrefix,
			"allowances",
			collections.PairKeyCodec(sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[feegrant.Grant](cdc),
		),
		FeeAllowanceQueue: collections.NewMap(
			sb,
			feegrant.FeeAllowanceQueueKeyPrefix,
			"allowances_queue",
			collections.TripleKeyCodec(sdk.TimeKey, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collections.BoolValue,
		),
	}
}

// GrantAllowance creates a new grant
func (k Keeper) GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error {
	granteeAddrStr, err := k.addrCdc.BytesToString(grantee)
	if err != nil {
		return err
	}
	exec, err := v1.NewMsgGrantAllowance(feeAllowance, granteeAddrStr)
	if err != nil {
		return err
	}
	_, err = k.accountKeeper.Execute(ctx, granter, granter, exec, nil)
	return err

}

// UpdateAllowance updates the existing grant.
func (k Keeper) UpdateAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error {
	panic("not supported")
}

func (k Keeper) revokeAllowance(ctx context.Context, granter, grantee sdk.AccAddress) error {
	panic("not supported")
}

// GetAllowance returns the allowance between the granter and grantee.
// If there is none, it returns nil, collections.ErrNotFound.
// Returns an error on parsing issues
func (k Keeper) GetAllowance(ctx context.Context, granter, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error) {
	grant, err := k.FeeAllowance.Get(ctx, collections.Join(grantee, granter))
	if err != nil {
		return nil, err
	}

	return grant.GetGrant()
}

// IterateAllFeeAllowances iterates over all the grants in the store.
// Callback to get all data, returns true to stop, false to keep reading
// Calling this without pagination is very expensive and only designed for export genesis
func (k Keeper) IterateAllFeeAllowances(ctx context.Context, cb func(grant feegrant.Grant) bool) error {
	return k.FeeAllowance.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], grant feegrant.Grant) (stop bool, err error) {
		return cb(grant), nil
	})
}

// UseGrantedFees will try to pay the given fee from the granter's account as requested by the grantee
func (k Keeper) UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
	granteeAddrStr, err := k.addrCdc.BytesToString(grantee)
	if err != nil {
		return err
	}
	exec, err := v1.NewMsgUseGrantedFees(granteeAddrStr, msgs...)
	if err != nil {
		return err
	}
	_, err = k.accountKeeper.Execute(ctx, granter, grantee, exec, nil)
	return err
}

func (k *Keeper) emitUseGrantEvent(ctx context.Context, granter, grantee string) error {
	return k.EventService.EventManager(ctx).EmitKV(
		feegrant.EventTypeUseFeeGrant,
		event.NewAttribute(feegrant.AttributeKeyGranter, granter),
		event.NewAttribute(feegrant.AttributeKeyGrantee, grantee),
	)
}

// InitGenesis will initialize the keeper from a *previously validated* GenesisState
func (k Keeper) InitGenesis(ctx context.Context, data *feegrant.GenesisState) error {
	for _, f := range data.Allowances {
		granter, err := k.addrCdc.StringToBytes(f.Granter)
		if err != nil {
			return err
		}
		grantee, err := k.addrCdc.StringToBytes(f.Grantee)
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
func (k Keeper) RemoveExpiredAllowances(ctx context.Context, limit int) error {
	exp := k.HeaderService.HeaderInfo(ctx).Time
	rng := collections.NewPrefixUntilTripleRange[time.Time, sdk.AccAddress, sdk.AccAddress](exp)
	count := 0

	keysToRemove := []collections.Triple[time.Time, sdk.AccAddress, sdk.AccAddress]{}
	err := k.FeeAllowanceQueue.Walk(ctx, rng, func(key collections.Triple[time.Time, sdk.AccAddress, sdk.AccAddress], value bool) (stop bool, err error) {
		grantee, granter := key.K2(), key.K3()

		if err := k.FeeAllowance.Remove(ctx, collections.Join(grantee, granter)); err != nil {
			return true, err
		}

		keysToRemove = append(keysToRemove, key)

		// limit the amount of iterations to avoid taking too much time
		count++
		if count == limit {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	for _, key := range keysToRemove {
		if err := k.FeeAllowanceQueue.Remove(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
