package common

import (
	"encoding/json"
	"fmt"
)

// Convenience struct for CLI output
type PrettyParams struct {
	CommunityTax        string `json:"community_tax"`
	BaseProposerReward  string `json:"base_proposer_reward"`
	BonusProposerReward string `json:"bonus_proposer_reward"`
	WithdrawAddrEnabled string `json:"withdraw_addr_enabled"`
}

// Construct a new PrettyParams
func NewPrettyParams(communityTax json.RawMessage, baseProposerReward json.RawMessage, bonusProposerReward json.RawMessage, withdrawAddrEnabled json.RawMessage) PrettyParams {
	return PrettyParams{
		CommunityTax:        string(communityTax),
		BaseProposerReward:  string(baseProposerReward),
		BonusProposerReward: string(bonusProposerReward),
		WithdrawAddrEnabled: string(withdrawAddrEnabled),
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
