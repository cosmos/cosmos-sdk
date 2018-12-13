package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// the address for where distributions rewards are withdrawn to by default
// this struct is only used at genesis to feed in default withdraw addresses
type DelegatorWithdrawInfo struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	WithdrawAddr  sdk.AccAddress `json:"withdraw_addr"`
}

// GenesisState - all distribution state that must be provided at genesis
type GenesisState struct {
	FeePool                FeePool                 `json:"fee_pool"`
	CommunityTax           sdk.Dec                 `json:"community_tax"`
	BaseProposerReward     sdk.Dec                 `json:"base_proposer_reward"`
	BonusProposerReward    sdk.Dec                 `json:"bonus_proposer_reward"`
	DelegatorWithdrawInfos []DelegatorWithdrawInfo `json:"delegator_withdraw_infos"`
	PreviousProposer       sdk.ConsAddress         `json:"previous_proposer"`
}

func NewGenesisState(feePool FeePool, communityTax, baseProposerReward, bonusProposerReward sdk.Dec, dwis []DelegatorWithdrawInfo, pp sdk.ConsAddress) GenesisState {

	return GenesisState{
		FeePool:                feePool,
		CommunityTax:           communityTax,
		BaseProposerReward:     baseProposerReward,
		BonusProposerReward:    bonusProposerReward,
		DelegatorWithdrawInfos: dwis,
		PreviousProposer:       pp,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		FeePool:             InitialFeePool(),
		CommunityTax:        sdk.NewDecWithPrec(2, 2), // 2%
		BaseProposerReward:  sdk.NewDecWithPrec(1, 2), // 1%
		BonusProposerReward: sdk.NewDecWithPrec(4, 2), // 4%
	}
}

// ValidateGenesis validates the genesis state of distribution genesis input
func ValidateGenesis(data GenesisState) error {
	if data.CommunityTax.IsNegative() || data.CommunityTax.GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameter CommunityTax should non-negative and "+
			"less than one, is %s", data.CommunityTax.String())
	}
	if data.BaseProposerReward.IsNegative() {
		return fmt.Errorf("mint parameter BaseProposerReward should be positive, is %s",
			data.BaseProposerReward.String())
	}
	if data.BonusProposerReward.IsNegative() {
		return fmt.Errorf("mint parameter BonusProposerReward should be positive, is %s",
			data.BonusProposerReward.String())
	}
	if (data.BaseProposerReward.Add(data.BonusProposerReward)).
		GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameters BaseProposerReward and "+
			"BonusProposerReward cannot add to be greater than one, "+
			"adds to %s", data.BaseProposerReward.Add(data.BonusProposerReward).String())
	}
	return data.FeePool.ValidateGenesis()
}
