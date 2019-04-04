package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Slashing tags
var (
	ActionValidatorUnjailed = "validator-unjailed"
	TxCategory              = "slashing"

	Action    = sdk.TagAction
	Category  = sdk.TagCategory
	Validator = "validator"
)
