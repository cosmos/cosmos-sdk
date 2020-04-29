package types

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
)

type (

	// AccountKeeper defines the expected account keeper interface used for simulations.
	AccountKeeper interface {
		GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.Account
	}

	// BankKeeper defines the expected bank keeper interface used for simulations.
	BankKeeper interface {
		GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
		GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
		SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	}

	// StakingKeeper defines the staking module interface contract needed by the
	// evidence module.
	StakingKeeper interface {
		ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingexported.ValidatorI
	}

	// SlashingKeeper defines the slashing module interface contract needed by the
	// evidence module.
	SlashingKeeper interface {
		GetPubkey(sdk.Context, crypto.Address) (crypto.PubKey, error)
		IsTombstoned(sdk.Context, sdk.ConsAddress) bool
		HasValidatorSigningInfo(sdk.Context, sdk.ConsAddress) bool
		Tombstone(sdk.Context, sdk.ConsAddress)
		Slash(sdk.Context, sdk.ConsAddress, sdk.Dec, int64, int64)
		SlashFractionDoubleSign(sdk.Context) sdk.Dec
		Jail(sdk.Context, sdk.ConsAddress)
		JailUntil(sdk.Context, sdk.ConsAddress, time.Time)
	}
)
