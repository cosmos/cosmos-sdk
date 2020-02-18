package vesting

// DONTCOVER
// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	RegisterCodec                  = types.RegisterCodec
	NewBaseVestingAccount          = types.NewBaseVestingAccount
	NewContinuousVestingAccountRaw = types.NewContinuousVestingAccountRaw
	NewContinuousVestingAccount    = types.NewContinuousVestingAccount
	NewPeriodicVestingAccountRaw   = types.NewPeriodicVestingAccountRaw
	NewPeriodicVestingAccount      = types.NewPeriodicVestingAccount
	NewDelayedVestingAccountRaw    = types.NewDelayedVestingAccountRaw
	NewDelayedVestingAccount       = types.NewDelayedVestingAccount
)

type (
	BaseVestingAccount       = types.BaseVestingAccount
	ContinuousVestingAccount = types.ContinuousVestingAccount
	PeriodicVestingAccount   = types.PeriodicVestingAccount
	DelayedVestingAccount    = types.DelayedVestingAccount
	Period                   = types.Period
	Periods                  = types.Periods
)
