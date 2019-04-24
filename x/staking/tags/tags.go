package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// staking tags
var (
	ActionCompleteUnbonding    = "complete-unbonding"
	ActionCompleteRedelegation = "complete-redelegation"
	TxCategory                 = "staking"

	Action       = sdk.TagAction
	Category     = sdk.TagCategory
	Sender       = sdk.TagSender
	SrcValidator = sdk.TagSrcValidator
	DstValidator = sdk.TagDstValidator
	Delegator    = sdk.TagDelegator
	EndTime      = "end-time"
)
