package types

import (
	context "context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected interface contract the vesting module
// requires for storing accounts.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
}

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	BlockedAddr(addr sdk.AccAddress) bool
}

// StakingKeeper defines the expected interface contract the vesting module
// requires for finding and changing the delegated tokens, used in clawback.
type StakingKeeper interface {
	BondDenom(ctx context.Context) string
	GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) math.Int
	GetDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress,
		maxRetrieve uint16) []stakingtypes.Delegation
	GetDelegatorUnbonding(ctx context.Context, delegator sdk.AccAddress) math.Int
	GetUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress,
		maxRetrieve uint16) []stakingtypes.UnbondingDelegation
	GetValidator(ctx context.Context, valAddr sdk.ValAddress) (stakingtypes.Validator, bool)
	TransferUnbonding(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantAmt math.Int) math.Int
	TransferDelegation(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantShares math.LegacyDec) math.LegacyDec
}

// DistributionHooks is the expected interface for distribution module hooks.
type DistributionHooks interface {
	// AllowWithdrawAddr tells whether to honor the delegation withdraw
	// address associated with the address (if any). The distribution
	// keeper will call this before each reward withdrawal.
	// If multiple distribution hooks are set, then any of them may
	// disallow the withdraw address.
	AllowWithdrawAddr(ctx context.Context, delAddr sdk.AccAddress) bool

	// AfterDelegationReward is called after the reward has been transferred the address.
	AfterDelegationReward(ctx context.Context, delAddr, withdrawAddr sdk.AccAddress, reward sdk.Coins)
}
