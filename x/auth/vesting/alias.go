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
	NewGenesisState                = types.NewGenesisState
	DefaultGenesisState            = types.DefaultGenesisState
	ValidateGenesis                = types.DefaultGenesisState
)

type (
	VestingAccount           = exported.VestingAccount
	BaseVestingAccount       = types.BaseVestingAccount
	ContinuousVestingAccount = types.ContinuousVestingAccount
	PeriodicVestingAccount   = types.PeriodicVestingAccount
	DelayedVestingAccount    = types.DelayedVestingAccount
	GenesisState             = types.GenesisState
	VestingPeriods           = types.VestingPeriods
	VestingPeriod            = types.VestingPeriod
)
