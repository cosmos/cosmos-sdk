package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	UnbondingTime(ctx sdk.Context) time.Duration
}

// UpgradeKeeper expected upgrade keeper
type UpgradeKeeper interface {
	GetUpgradePlan(ctx sdk.Context) (plan upgradetypes.Plan, havePlan bool)
	SetUpgradedConsensusState(ctx sdk.Context, planHeight int64, bz []byte) error
}
