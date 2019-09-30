package vesting

import (
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
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
)

type (
	Account                  = exported.VestingAccount
	BaseVestingAccount       = types.BaseVestingAccount
	ContinuousVestingAccount = types.ContinuousVestingAccount
	PeriodicVestingAccount   = types.PeriodicVestingAccount
	DelayedVestingAccount    = types.DelayedVestingAccount
	Periods                  = types.Periods
	Period                   = types.Period
)
