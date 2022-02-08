package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected interface contract the vesting module
// requires for storing accounts.
type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	BlockedAddr(addr sdk.AccAddress) bool
}

// StakingKeeper defines the expected interface contract the vesting module
// requires for finding and changing the delegated tokens, used in clawback.
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) sdk.Int
	GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		maxRetrieve uint16) []stakingtypes.Delegation
	GetDelegatorUnbonding(ctx sdk.Context, delegator sdk.AccAddress) sdk.Int
	GetUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		maxRetrieve uint16) []stakingtypes.UnbondingDelegation
	GetValidator(ctx sdk.Context, valAddr sdk.ValAddress) (stakingtypes.Validator, bool)
	TransferUnbonding(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantAmt sdk.Int) sdk.Int
	TransferDelegation(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantShares sdk.Dec) sdk.Dec
}

// DistributionHooks is the expected interface for distribution module hooks.
type DistributionHooks interface {
	// AllowWithdrawAddr tells whether to honor the delegation withdraw
	// address associated with the address (if any). The distribution
	// keeper will call this before each reward withdrawal.
	// If multiple distribution hooks are set, then any of them may
	// disallow the withdraw address.
	AllowWithdrawAddr(ctx sdk.Context, delAddr sdk.AccAddress) bool

	// AfterDelegationReward is called after the reward has been transferred the address.
	AfterDelegationReward(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress, reward sdk.Coins)
}
