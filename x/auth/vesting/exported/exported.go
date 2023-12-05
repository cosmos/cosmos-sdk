package exported

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VestingAccount defines an account type that vests coins via a vesting schedule.
type VestingAccount interface {
	sdk.AccountI

	// LockedCoins returns the set of coins that are not spendable (i.e. locked),
	// defined as the vesting coins that are not delegated.
	//
	// To get spendable coins of a vesting account, first the total balance must
	// be retrieved and the locked tokens can be subtracted from the total balance.
	// Note, the spendable balance can be negative.
	LockedCoins(ctx sdk.Context) sdk.Coins

	// TrackDelegation performs internal vesting accounting necessary when
	// delegating from a vesting account. It accepts the current block time, the
	// delegation amount and balance of all coins whose denomination exists in
	// the account's original vesting balance.
	TrackDelegation(blockTime time.Time, balance, amount sdk.Coins)

	// TrackUndelegation performs internal vesting accounting necessary when a
	// vesting account performs an undelegation.
	TrackUndelegation(amount sdk.Coins)

	GetVestedCoins(blockTime time.Time) sdk.Coins
	GetVestingCoins(blockTime time.Time) sdk.Coins

	GetStartTime() int64
	GetEndTime() int64

	GetOriginalVesting() sdk.Coins
	GetDelegatedFree() sdk.Coins
	GetDelegatedVesting() sdk.Coins
}

// Additional vesting account behaviors are abstracted by interfaces to avoid
// cyclic package dependencies. Specifically, the account-wrapping mechanism
// to support liens in Agoric/agoric-sdk.

// AddGrantAction encapsulates the data needed to add a grant to an account.
type AddGrantAction interface {
	// AddToAccount adds the grant to the specified account.
	// The rawAccount should bypass any account wrappers.
	AddToAccount(ctx sdk.Context, rawAccount VestingAccount) error
}

// ReturnGrantAction encapsulates the data needed to return grants from an account.
type ReturnGrantAction interface {
	// TakeGrants removes the original vesting amount from the account
	// and clears the original vesting amount and schedule.
	// The rawAccount should bypass any account wrappers.
	TakeGrants(ctx sdk.Context, rawAccount VestingAccount) error
}

// ClabackAction encapsulates the data needed to perform clawback.
type ClawbackAction interface {
	// TakeFromAccount removes unvested tokens from the specified account.
	// The rawAccount should bypass any account wrappers.
	TakeFromAccount(ctx sdk.Context, rawAccount VestingAccount) error
}

// RewardAction encapsulates the data needed to process rewards distributed to an account.
type RewardAction interface {
	// ProcessReward processes the given reward which as been added to the account.
	// Returns an error if the account is of the wrong type.
	// The rawAccount should bypass any account wrappers.
	ProcessReward(ctx sdk.Context, reward sdk.Coins, rawAccount VestingAccount) error
}

// GrantAccount is a VestingAccount which can accept new grants.
type GrantAccount interface {
	VestingAccount
	// AddGrant adds the specified grant to the account.
	AddGrant(ctx sdk.Context, action AddGrantAction) error
}

// ClawbackVestingAccountI is an interface for the methods of a clawback account.
type ClawbackVestingAccountI interface {
	GrantAccount

	// GetUnlockedOnly returns the sum of all unlocking events up to and including
	// the blockTime.
	GetUnlockedOnly(blockTime time.Time) sdk.Coins

	// GetVestedOnly returns the sum of all vesting events up to and including
	// the blockTime.
	GetVestedOnly(blockTime time.Time) sdk.Coins

	// Clawback performs the clawback described by action.
	Clawback(ctx sdk.Context, action ClawbackAction) error

	// PostReward preforms post-reward processing described by action.
	PostReward(ctx sdk.Context, reward sdk.Coins, action RewardAction) error

	// ReturnGrants returns all grants to the funder.
	ReturnGrants(ctx sdk.Context, action ReturnGrantAction) error
}
