package cli

import (
	"encoding/json"
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
