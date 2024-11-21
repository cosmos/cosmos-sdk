package lockup

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	assettypes "cosmossdk.io/x/accounts/defaults/asset/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	OriginalLockingPrefix  = collections.NewPrefix(0)
	DelegatedFreePrefix    = collections.NewPrefix(1)
	DelegatedLockingPrefix = collections.NewPrefix(2)
	EndTimePrefix          = collections.NewPrefix(3)
	StartTimePrefix        = collections.NewPrefix(4)
	LockingPeriodsPrefix   = collections.NewPrefix(5)
	OwnerPrefix            = collections.NewPrefix(6)
	WithdrawedCoinsPrefix  = collections.NewPrefix(7)
)

var (
	CONTINUOUS_LOCKING_ACCOUNT = "continuous-locking-account"
	DELAYED_LOCKING_ACCOUNT    = "delayed-locking-account"
	PERIODIC_LOCKING_ACCOUNT   = "periodic-locking-account"
	PERMANENT_LOCKING_ACCOUNT  = "permanent-locking-account"
)

type getLockedCoinsFunc = func(ctx context.Context, time time.Time, denoms ...string) (sdk.Coins, error)

type bankKeeperI interface {
	
}

// newBaseLockup creates a new BaseLockup object.
func newBaseLockup(d accountstd.Dependencies) *BaseLockup {
	BaseLockup := &BaseLockup{
		Owner:            collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		OriginalLocking:  collections.NewMap(d.SchemaBuilder, OriginalLockingPrefix, "original_locking", collections.StringKey, sdk.IntValue),
		DelegatedFree:    collections.NewMap(d.SchemaBuilder, DelegatedFreePrefix, "delegated_free", collections.StringKey, sdk.IntValue),
		DelegatedLocking: collections.NewMap(d.SchemaBuilder, DelegatedLockingPrefix, "delegated_locking", collections.StringKey, sdk.IntValue),
		WithdrawedCoins:  collections.NewMap(d.SchemaBuilder, WithdrawedCoinsPrefix, "withdrawed_coins", collections.StringKey, sdk.IntValue),
		addressCodec:     d.AddressCodec,
		headerService:    d.Environment.HeaderService,
		EndTime:          collections.NewItem(d.SchemaBuilder, EndTimePrefix, "end_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
	}

	return BaseLockup
}

type BaseLockup struct {
	// Owner is the address of the account owner.
	Owner            collections.Item[[]byte]
	OriginalLocking  collections.Map[string, math.Int]
	DelegatedFree    collections.Map[string, math.Int]
	DelegatedLocking collections.Map[string, math.Int]
	WithdrawedCoins  collections.Map[string, math.Int]
	addressCodec     address.Codec
	headerService    header.Service
	// lockup end time.
	EndTime collections.Item[time.Time]
}

func (bva *BaseLockup) Init(ctx context.Context, msg *lockuptypes.MsgInitLockupAccount) (
	*lockuptypes.MsgInitLockupAccountResponse, error,
) {
	owner, err := bva.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	err = bva.Owner.Set(ctx, owner)
	if err != nil {
		return nil, err
	}

	funds := accountstd.Funds(ctx)

	sortedAmt := funds.Sort()
	for _, coin := range sortedAmt {
		err = bva.OriginalLocking.Set(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return nil, err
		}

		// Set initial value for all locked token
		err = bva.WithdrawedCoins.Set(ctx, coin.Denom, math.ZeroInt())
		if err != nil {
			return nil, err
		}
	}

	bondDenom, err := getStakingDenom(ctx)
	if err != nil {
		return nil, err
	}

	// Set initial value for all locked token
	err = bva.DelegatedFree.Set(ctx, bondDenom, math.ZeroInt())
	if err != nil {
		return nil, err
	}

	// Set initial value for all locked token
	err = bva.DelegatedLocking.Set(ctx, bondDenom, math.ZeroInt())
	if err != nil {
		return nil, err
	}

	err = bva.EndTime.Set(ctx, msg.EndTime)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgInitLockupAccountResponse{}, nil
}
