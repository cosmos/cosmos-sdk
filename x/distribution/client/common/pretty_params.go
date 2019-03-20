package common

import (
	"encoding/json"
	"fmt"
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
