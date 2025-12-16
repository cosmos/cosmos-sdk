package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultNakamotoBonus is the ADR's initial value: 3% (0.03)
var DefaultNakamotoBonus = math.LegacyNewDecWithPrec(3, 2) // 0.03

func NewGenesisState(
	params Params,
	fp FeePool,
	dwis []DelegatorWithdrawInfo,
	pp sdk.ConsAddress,
	r []ValidatorOutstandingRewardsRecord,
	acc []ValidatorAccumulatedCommissionRecord,
	historical []ValidatorHistoricalRewardsRecord,
	cur []ValidatorCurrentRewardsRecord,
	dels []DelegatorStartingInfoRecord,
	slashes []ValidatorSlashEventRecord,
	nakamotoBonus math.LegacyDec,
) *GenesisState {
	return &GenesisState{
		Params:                          params,
		FeePool:                         fp,
		DelegatorWithdrawInfos:          dwis,
		PreviousProposer:                pp.String(),
		OutstandingRewards:              r,
		ValidatorAccumulatedCommissions: acc,
		ValidatorHistoricalRewards:      historical,
		ValidatorCurrentRewards:         cur,
		DelegatorStartingInfos:          dels,
		ValidatorSlashEvents:            slashes,
		NakamotoBonus:                   nakamotoBonus,
	}
}

// DefaultGenesisState get raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		FeePool:                         InitialFeePool(),
		Params:                          DefaultParams(),
		DelegatorWithdrawInfos:          []DelegatorWithdrawInfo{},
		PreviousProposer:                "",
		OutstandingRewards:              []ValidatorOutstandingRewardsRecord{},
		ValidatorAccumulatedCommissions: []ValidatorAccumulatedCommissionRecord{},
		ValidatorHistoricalRewards:      []ValidatorHistoricalRewardsRecord{},
		ValidatorCurrentRewards:         []ValidatorCurrentRewardsRecord{},
		DelegatorStartingInfos:          []DelegatorStartingInfoRecord{},
		ValidatorSlashEvents:            []ValidatorSlashEventRecord{},
		NakamotoBonus:                   DefaultNakamotoBonus,
	}
}

// ValidateGenesis validates the genesis state of distribution genesis input
func ValidateGenesis(gs *GenesisState) error {
	if gs.NakamotoBonus.IsNegative() || gs.NakamotoBonus.GT(math.LegacyOneDec()) {
		return fmt.Errorf("nakamoto bonus must be within [0,1], got %s", gs.NakamotoBonus.String())
	}

	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}
	return gs.FeePool.ValidateGenesis()
}
