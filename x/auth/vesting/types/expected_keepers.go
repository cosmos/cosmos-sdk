package types

import (
	context "context"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BankKeeper defines the expected interface contract the vesting module requires
// for creating vesting accounts with funds.
type BankKeeper interface {
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	BlockedAddr(addr sdk.AccAddress) bool
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected interface contract the vesting module
// requires for storing accounts.
type AccountKeeper interface {
	// Retrieve an account from the store.
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI

	// Set an account in the store.
	SetAccount(context.Context, sdk.AccountI)
}

// StakingKeeper defines the expected interface contract the vesting module
// requires for finding and changing the delegated tokens, used in clawback.
type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
	GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (math.Int, error)
	GetDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress,
		maxRetrieve uint16) ([]stakingtypes.Delegation, error)
	GetDelegatorUnbonding(ctx context.Context, delegator sdk.AccAddress) (math.Int, error)
	GetUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress,
		maxRetrieve uint16) ([]stakingtypes.UnbondingDelegation, error)
	GetValidator(ctx context.Context, valAddr sdk.ValAddress) (stakingtypes.Validator, error)
	TransferUnbonding(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantAmt math.Int) (math.Int, error)
	TransferDelegation(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantShares math.LegacyDec) (math.LegacyDec, error)
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
