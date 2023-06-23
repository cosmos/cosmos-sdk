package types

// staking module event types
const (
	EventTypeCompleteUnbonding           = "complete_unbonding"
	EventTypeCompleteRedelegation        = "complete_redelegation"
	EventTypeCreateValidator             = "create_validator"
	EventTypeEditValidator               = "edit_validator"
	EventTypeDelegate                    = "delegate"
	EventTypeUnbond                      = "unbond"
	EventTypeRedelegate                  = "redelegate"
	EventTypeTokenizeShares              = "tokenize_shares"
	EventTypeRedeemShares                = "redeem_shares"
	EventTypeTransferTokenizeShareRecord = "transfer_tokenize_share_record"
	EventTypeValidatorBondDelegation     = "validator_bond_delegation"

	AttributeKeyValidator      = "validator"
	AttributeKeyCommissionRate = "commission_rate"
	AttributeKeySrcValidator   = "source_validator"
	AttributeKeyDstValidator   = "destination_validator"
	AttributeKeyDelegator      = "delegator"
	AttributeKeyCompletionTime = "completion_time"
	AttributeKeyNewShares      = "new_shares"
	AttributeKeyShareOwner     = "share_owner"
	AttributeKeyShareRecordID  = "share_record_id"
	AttributeKeyAmount         = "amount"
	AttributeValueCategory     = ModuleName
)
