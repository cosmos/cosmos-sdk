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
	ValidatorDistInfos     []ValidatorDistInfo     `json:"validator_dist_infos"`
	DelegatorDistInfos     []DelegatorDistInfo     `json:"delegator_dist_infos"`
	DelegatorWithdrawInfos []DelegatorWithdrawInfo `json:"delegator_withdraw_infos"`
}

func NewGenesisState(feePool FeePool, vdis []ValidatorDistInfo, ddis []DelegatorDistInfo) GenesisState {
	return GenesisState{
		FeePool:            feePool,
		ValidatorDistInfos: vdis,
		DelegatorDistInfos: ddis,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		FeePool: InitialFeePool(),
	}
}
