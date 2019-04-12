package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// QueryDelegatorTotalRewardsResponse defines the properties of
// QueryDelegatorTotalRewards query's response.
type QueryDelegatorTotalRewardsResponse struct {
	Rewards []DelegationDelegatorReward `json:"rewards"`
	Total   sdk.DecCoins                `json:"total"`
}

// NewQueryDelegatorTotalRewardsResponse constructs a QueryDelegatorTotalRewardsResponse
func NewQueryDelegatorTotalRewardsResponse(rewards []DelegationDelegatorReward,
	total sdk.DecCoins) QueryDelegatorTotalRewardsResponse {
	return QueryDelegatorTotalRewardsResponse{Rewards: rewards, Total: total}
}

// DelegationDelegatorReward defines the properties
// of a delegator's delegation reward.
type DelegationDelegatorReward struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
	Reward           sdk.DecCoins   `json:"reward"`
}

// NewDelegationDelegatorReward constructs a DelegationDelegatorReward.
func NewDelegationDelegatorReward(valAddr sdk.ValAddress,
	reward sdk.DecCoins) DelegationDelegatorReward {
	return DelegationDelegatorReward{ValidatorAddress: valAddr, Reward: reward}
}
