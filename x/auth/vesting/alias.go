package vesting

import (
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/internal/types"
)

var (
	NewBaseVestingAccount          = types.NewBaseVestingAccount
	NewContinuousVestingAccountRaw = types.NewContinuousVestingAccountRaw
	NewContinuousVestingAccount    = types.NewContinuousVestingAccount
	NewPeriodicVestingAccountRaw   = types.NewPeriodicVestingAccountRaw
	NewPeriodicVestingAccount      = types.NewPeriodicVestingAccount
	NewDelayedVestingAccountRaw    = types.NewDelayedVestingAccountRaw
	NewDelayedVestingAccount       = types.NewDelayedVestingAccount
	RegisterCodec                  = types.RegisterCodec
	RandomGenesisAccounts          = types.RandomGenesisAccounts
)

type (
	VestingAccount           = exported.VestingAccount
	BaseVestingAccount       = types.BaseVestingAccount
	ContinuousVestingAccount = types.ContinuousVestingAccount
	PeriodicVestingAccount   = types.PeriodicVestingAccount
	DelayedVestingAccount    = types.DelayedVestingAccount
	VestingPeriods           = types.VestingPeriods
	VestingPeriod            = types.VestingPeriod
)
