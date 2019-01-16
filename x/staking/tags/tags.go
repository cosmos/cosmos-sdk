// nolint
package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ActionCompleteUnbonding    = []byte("complete-unbonding")
	ActionCompleteRedelegation = []byte("complete-redelegation")

	Action       = sdk.TagAction
	SrcValidator = sdk.TagSrcValidator
	DstValidator = sdk.TagDstValidator
	Delegator    = sdk.TagDelegator
	Moniker      = "moniker"
	Identity     = "identity"
	EndTime      = "end-time"
)
