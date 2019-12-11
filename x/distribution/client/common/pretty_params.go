package common

import (
	"encoding/json"
	"fmt"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Convenience struct for CLI output
type PrettyParams struct {
	CommunityTax        json.RawMessage `json:"community_tax"`
	BaseProposerReward  json.RawMessage `json:"base_proposer_reward"`
	BonusProposerReward json.RawMessage `json:"bonus_proposer_reward"`
	WithdrawAddrEnabled json.RawMessage `json:"withdraw_addr_enabled"`
}

// Construct a new PrettyParams
func NewPrettyParams(communityTax json.RawMessage, baseProposerReward json.RawMessage, bonusProposerReward json.RawMessage, withdrawAddrEnabled json.RawMessage) PrettyParams {
	return PrettyParams{
		CommunityTax:        communityTax,
		BaseProposerReward:  baseProposerReward,
		BonusProposerReward: bonusProposerReward,
		WithdrawAddrEnabled: withdrawAddrEnabled,
	}
}

func (pp PrettyParams) String() string {
	return fmt.Sprintf(`Distribution Params:
  Community Tax:          %s
  Base Proposer Reward:   %s
  Bonus Proposer Reward:  %s
  Withdraw Addr Enabled:  %s`, pp.CommunityTax,
		pp.BaseProposerReward, pp.BonusProposerReward, pp.WithdrawAddrEnabled)

}
func (pp PrettyParams) MarshalYAML() (interface{}, error) {

	var CommunityTax sdk.Dec
	var BaseProposerReward sdk.Dec
	var BonusProposerReward sdk.Dec
	var WithdrawAddrEnabled bool

	err := json.Unmarshal(pp.CommunityTax, &CommunityTax)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(pp.BaseProposerReward, &BaseProposerReward)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(pp.BonusProposerReward, &BonusProposerReward)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(pp.WithdrawAddrEnabled, &WithdrawAddrEnabled)
	if err != nil {
		return nil, err
	}

	bz, err := yaml.Marshal(
		struct {
			CommunityTax        sdk.Dec
			BaseProposerReward  sdk.Dec
			BonusProposerReward sdk.Dec
			WithdrawAddrEnabled bool
		}{
			CommunityTax:        CommunityTax,
			BaseProposerReward:  BaseProposerReward,
			BonusProposerReward: BonusProposerReward,
			WithdrawAddrEnabled: WithdrawAddrEnabled,
		})
	if err != nil {
		return nil, err
	}
	return string(bz), nil
}
