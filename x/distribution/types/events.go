package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress     = "set_withdraw_address"
	EventTypeRewards                = "rewards"
	EventTypeCommission             = "commission"
	EventTypeWithdrawRewards        = "withdraw_rewards"
	EventTypeWithdrawCommission     = "withdraw_commission"
	EventTypeProposerReward         = "proposer_reward"
	EventTypeWithdrawAddrBlocked    = "withdraw_addr_blocked"
	EventTypeWithdrawAddrRedirected = "withdraw_addr_redirected"

	AttributeKeyWithdrawAddress         = "withdraw_address"
	AttributeKeyValidator               = "validator"
	AttributeKeyDelegator               = "delegator"
	AttributeKeyOriginalWithdrawAddress = "original_withdraw_address"

	AttributeValueCommunityPool = "community_pool"
)
