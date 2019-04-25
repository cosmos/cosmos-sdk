package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Crisis module tags
var (
	TxCategory = "crisis"

	Sender    = sdk.TagSender
	Category  = sdk.TagCategory
	Invariant = "invariant"
)
