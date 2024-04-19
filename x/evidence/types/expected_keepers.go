package types

import (
	"context"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper defines the staking module interface contract needed by the
// evidence module.
type StakingKeeper interface {
	ConsensusAddressCodec() address.Codec
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.ValidatorI, error)
	GetParams(ctx context.Context) (params stakingtypes.Params, err error)
}

// SlashingKeeper defines the slashing module interface contract needed by the
// evidence module.
type SlashingKeeper interface {
	GetPubkey(context.Context, cryptotypes.Address) (cryptotypes.PubKey, error)
	IsTombstoned(context.Context, sdk.ConsAddress) bool
	HasValidatorSigningInfo(context.Context, sdk.ConsAddress) bool
	Tombstone(context.Context, sdk.ConsAddress) error
	Slash(context.Context, sdk.ConsAddress, math.LegacyDec, int64, int64) error
	SlashWithInfractionReason(context.Context, sdk.ConsAddress, math.LegacyDec, int64, int64, stakingtypes.Infraction) error
	SlashFractionDoubleSign(context.Context) (math.LegacyDec, error)
	Jail(context.Context, sdk.ConsAddress) error
	JailUntil(context.Context, sdk.ConsAddress, time.Time) error
}

type Cometinfo interface {
	comet.BlockInfoService
}
