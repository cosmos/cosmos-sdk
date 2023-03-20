package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress = "set_withdraw_address"
	EventTypeRewards            = "rewards"
	EventTypeCommission         = "commission"
	EventTypeWithdrawRewards    = "withdraw_rewards"
	EventTypeWithdrawCommission = "withdraw_commission"
	EventTypeProposerReward     = "proposer_reward"

	AttributeKeyWithdrawAddress = "withdraw_address"
	AttributeKeyValidator       = "validator"
<<<<<<< HEAD

	AttributeValueCategory = ModuleName
=======
	AttributeKeyDelegator       = "delegator"
>>>>>>> 176c61cf7 (feat: add delegator to withdraw address (#15462))
)
