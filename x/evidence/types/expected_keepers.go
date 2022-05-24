package types

import (
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type (
	// StakingKeeper defines the staking module interface contract needed by the
	// evidence module.
	StakingKeeper interface {
		ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingtypes.ValidatorI
	}

	// SlashingKeeper defines the slashing module interface contract needed by the
	// evidence module.
	SlashingKeeper interface {
		GetPubkey(sdk.Context, cryptotypes.Address) (cryptotypes.PubKey, error)
		IsTombstoned(sdk.Context, sdk.ConsAddress) bool
		HasValidatorSigningInfo(sdk.Context, sdk.ConsAddress) bool
		Tombstone(sdk.Context, sdk.ConsAddress)
		Slash(sdk.Context, sdk.ConsAddress, sdk.Dec, int64, int64)
		SlashFractionDoubleSign(sdk.Context) sdk.Dec
		Jail(sdk.Context, sdk.ConsAddress)
		JailUntil(sdk.Context, sdk.ConsAddress, time.Time)
	}
)
