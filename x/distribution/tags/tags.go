// nolint
package tags

import (
	sdk "github.com/YunSuk-Yeo/cosmos-sdk/types"
)

// Distribution tx tags
var (
	Rewards    = "rewards"
	Commission = "commission"

	Validator = sdk.TagSrcValidator
	Delegator = sdk.TagDelegator
)
