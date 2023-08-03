package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress          = "set_withdraw_address"
	EventTypeRewards                     = "rewards"
	EventTypeCommission                  = "commission"
	EventTypeWithdrawRewards             = "withdraw_rewards"
	EventTypeWithdrawCommission          = "withdraw_commission"
	EventTypeWithdrawTokenizeShareReward = "withdraw_tokenize_share_reward"
	EventTypeProposerReward              = "proposer_reward"

	AttributeKeyWithdrawAddress = "withdraw_address"
	AttributeKeyValidator       = "validator"

	AttributeValueCategory = ModuleName
)
