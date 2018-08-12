// nolint
package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ActionCreateValidator      = []byte("create-validator")
	ActionEditValidator        = []byte("edit-validator")
	ActionDelegate             = []byte("delegate")
	ActionBeginUnbonding       = []byte("begin-unbonding")
	ActionCompleteUnbonding    = []byte("complete-unbonding")
	ActionBeginRedelegation    = []byte("begin-redelegation")
	ActionCompleteRedelegation = []byte("complete-redelegation")

	Action       = sdk.TagAction
	SrcValidator = sdk.TagSrcValidator
	DstValidator = sdk.TagDstValidator
	Delegator    = sdk.TagDelegator
	Moniker      = "moniker"
	Identity     = "identity"
)
