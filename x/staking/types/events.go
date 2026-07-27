package types

// staking module event types
const (
	EventTypeCompleteUnbonding         = "complete_unbonding"
	EventTypeCompleteRedelegation      = "complete_redelegation"
	EventTypeCreateValidator           = "create_validator"
	EventTypeEditValidator             = "edit_validator"
	EventTypeDelegate                  = "delegate"
	EventTypeUnbond                    = "unbond"
	EventTypeCancelUnbondingDelegation = "cancel_unbonding_delegation"
	EventTypeRedelegate                = "redelegate"
	EventTypeRotateConsPubKey          = "rotate_cons_pubkey"
	EventTypeApplyConsPubKeyRotation   = "apply_cons_pubkey_rotation"

	AttributeKeyValidator            = "validator"
	AttributeKeyCommissionRate       = "commission_rate"
	AttributeKeyMinSelfDelegation    = "min_self_delegation"
	AttributeKeySrcValidator         = "source_validator"
	AttributeKeyDstValidator         = "destination_validator"
	AttributeKeyDelegator            = "delegator"
	AttributeKeyCreationHeight       = "creation_height"
	AttributeKeyCompletionTime       = "completion_time"
	AttributeKeyNewShares            = "new_shares"
	AttributeKeyOldConsAddr          = "old_cons_addr"
	AttributeKeyNewConsAddr          = "new_cons_addr"
	AttributeKeyApplyHeight          = "apply_height"
	AttributeKeyMaturityTime         = "maturity_time"
	AttributeKeyEvidenceExpiryTime   = "evidence_expiry_time"
	AttributeKeyEvidenceExpiryHeight = "evidence_expiry_height"
	AttributeKeyFeeBurned            = "fee_burned"
)
