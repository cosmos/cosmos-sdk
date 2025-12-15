package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress = "set_withdraw_address"
	EventTypeRewards            = "rewards"
	EventTypeCommission         = "commission"
	EventTypeWithdrawRewards    = "withdraw_rewards"
	EventTypeWithdrawCommission = "withdraw_commission"

	AttributeKeyWithdrawAddress = "withdraw_address"
	AttributeKeyValidator       = "validator"
	AttributeKeyDelegator       = "delegator"
)
