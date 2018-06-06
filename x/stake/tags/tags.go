// nolint
package tags

import (
	"github.com/cosmos/cosmos-sdk/types"
)

var (
	ActionCreateValidator      = []byte("create-validator")
	ActionEditValidator        = []byte("edit-validator")
	ActionDelegate             = []byte("delegate")
	ActionBeginUnbonding       = []byte("begin-unbonding")
	ActionCompleteUnbonding    = []byte("complete-unbonding")
	ActionBeginRedelegation    = []byte("begin-redelegation")
	ActionCompleteRedelegation = []byte("complete-redelegation")

	Action       = types.TagAction
	SrcValidator = types.TagSrcValidator
	DstValidator = types.TagDstValidator
	Delegator    = types.TagDelegator
	Slashed      = []byte("slashed")
	Moniker      = []byte("moniker")
	Identity     = []byte("Identity")
)
