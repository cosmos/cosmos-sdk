package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	TxCategory = "distribution"

	Validator = sdk.TagSrcValidator
	Delegator = sdk.TagDelegator
	Category  = sdk.TagCategory
)
