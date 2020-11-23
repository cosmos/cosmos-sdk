package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	GetLastValidators(ctx sdk.Context) []stakingtypes.Validator
	UnbondingTime(ctx sdk.Context) time.Duration
}
