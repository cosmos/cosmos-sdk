package types

import (
	"cosmossdk.io/collections"
)

var (
	OriginalLockingPrefix  = collections.NewPrefix(0)
	DelegatedFreePrefix    = collections.NewPrefix(1)
	DelegatedLockingPrefix = collections.NewPrefix(2)
	EndTimePrefix          = collections.NewPrefix(3)
	StartTimePrefix        = collections.NewPrefix(4)
	LockingPeriodsPrefix   = collections.NewPrefix(5)
	OwnerPrefix            = collections.NewPrefix(6)
	WithdrawedCoinsPrefix  = collections.NewPrefix(7)
	AdminPrefix            = collections.NewPrefix(8)
	OriginalVestingPrefix  = collections.NewPrefix(9)
)

var (
	CONTINUOUS_LOCKING_ACCOUNT = "continuous"
	DELAYED_LOCKING_ACCOUNT    = "delayed"
	PERIODIC_LOCKING_ACCOUNT   = "periodic"
	PERMANENT_LOCKING_ACCOUNT  = "permanent"
	CLAWBACK_ENABLE_SUFFIX     = "-clawback-enable"
)
