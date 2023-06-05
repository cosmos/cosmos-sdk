package types

import (
	"context"
	"time"

	"cosmossdk.io/core/comet"
	sdkmath "cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type (
	// StakingKeeper defines the staking module interface contract needed by the
	// evidence module.
	StakingKeeper interface {
		ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingtypes.ValidatorI
		GetParams(ctx sdk.Context) (params stakingtypes.Params)
	}

	// SlashingKeeper defines the slashing module interface contract needed by the
	// evidence module.
	SlashingKeeper interface {
		GetPubkey(context.Context, cryptotypes.Address) (cryptotypes.PubKey, error)
		IsTombstoned(context.Context, sdk.ConsAddress) bool
		HasValidatorSigningInfo(context.Context, sdk.ConsAddress) bool
		Tombstone(context.Context, sdk.ConsAddress) error
		Slash(context.Context, sdk.ConsAddress, sdkmath.LegacyDec, int64, int64) error
		SlashWithInfractionReason(context.Context, sdk.ConsAddress, sdkmath.LegacyDec, int64, int64, stakingtypes.Infraction) error
		SlashFractionDoubleSign(context.Context) (sdkmath.LegacyDec, error)
		Jail(context.Context, sdk.ConsAddress) error
		JailUntil(context.Context, sdk.ConsAddress, time.Time) error
	}

	// AccountKeeper define the account keeper interface contracted needed by the evidence module
	AccountKeeper interface {
		SetAccount(ctx context.Context, acc sdk.AccountI)
	}

	// BankKeeper define the account keeper interface contracted needed by the evidence module
	BankKeeper interface {
		MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
		SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
		GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	}

	Cometinfo interface {
		comet.BlockInfoService
	}
)
