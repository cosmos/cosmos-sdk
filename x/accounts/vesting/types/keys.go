package types

import "cosmossdk.io/collections"

var (
	OriginalVestingPrefix  = collections.NewPrefix(0)
	DelegatedFreePrefix    = collections.NewPrefix(1)
	DelegatedVestingPrefix = collections.NewPrefix(2)
	EndTimePrefix          = collections.NewPrefix(3)
	StartTimePrefix        = collections.NewPrefix(4)
	VestingPeriodsPrefix   = collections.NewPrefix(5)
	OwnerPrefix            = collections.NewPrefix(6)
)

var (
	CONTINUOUS_VESTING_ACCOUNT = "continuous-vesting-account"
	DELAYED_VESTING_ACCOUNT    = "delayed-vesting-account"
	PERIODIC_VESTING_ACCOUNT   = "periodic-vesting-account"
	PERMANENT_VESTING_ACCOUNT  = "permanent-vesting-account"
)
