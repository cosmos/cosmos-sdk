package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	TxCategory = "distribution"

	Validator = sdk.TagSrcValidator
	Category  = sdk.TagCategory
	Sender    = sdk.TagSender
)
