package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Slashing tags
var (
	ActionValidatorUnjailed = []byte("validator-unjailed")

	Action    = sdk.TagAction
	Validator = "validator"
)
