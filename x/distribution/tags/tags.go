// nolint
package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Distribution tx tags
var (
	Rewards    = "rewards"
	Commission = "commission"

	Validator = sdk.TagSrcValidator
	Delegator = sdk.TagDelegator
)
