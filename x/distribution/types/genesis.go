package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	ValidatorDistInfos     []ValidatorDistInfo     `json:"validator_dist_infos"`
	DelegationDistInfos    []DelegationDistInfo    `json:"delegator_dist_infos"`
	DelegatorWithdrawInfos []DelegatorWithdrawInfo `json:"delegator_withdraw_infos"`
	PreviousProposer       sdk.ConsAddress         `json:"previous_proposer"`
}

func NewGenesisState(feePool FeePool, communityTax, baseProposerReward, bonusProposerReward sdk.Dec,
	vdis []ValidatorDistInfo, ddis []DelegationDistInfo, dwis []DelegatorWithdrawInfo, pp sdk.ConsAddress) GenesisState {

	return GenesisState{
		FeePool:                feePool,
		CommunityTax:           communityTax,
		BaseProposerReward:     baseProposerReward,
		BonusProposerReward:    bonusProposerReward,
		ValidatorDistInfos:     vdis,
		DelegationDistInfos:    ddis,
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

// default genesis utility function, initialize for starting validator set
func DefaultGenesisWithValidators(valAddrs []sdk.ValAddress) GenesisState {

	vdis := make([]ValidatorDistInfo, len(valAddrs))
	ddis := make([]DelegationDistInfo, len(valAddrs))

	for i, valAddr := range valAddrs {
		vdis[i] = NewValidatorDistInfo(valAddr, 0)
		accAddr := sdk.AccAddress(valAddr)
		ddis[i] = NewDelegationDistInfo(accAddr, valAddr, 0)
	}

	return GenesisState{
		FeePool:             InitialFeePool(),
		CommunityTax:        sdk.NewDecWithPrec(2, 2), // 2%
		BaseProposerReward:  sdk.NewDecWithPrec(1, 2), // 1%
		BonusProposerReward: sdk.NewDecWithPrec(4, 2), // 4%
		ValidatorDistInfos:  vdis,
		DelegationDistInfos: ddis,
	}
}
