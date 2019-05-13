// nolint
package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Distribution tx tags
var (
	TxCategory = "distribution"

	Rewards    = "rewards"
	Commission = "commission"
	Validator = sdk.TagSrcValidator
	Category  = sdk.TagCategory
	Sender    = sdk.TagSender
)
