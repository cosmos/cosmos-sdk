package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Tag keys and values
var (
	Category   = sdk.EventTypeCategory
	TxCategory = "bank"
	Transfer   = "transfer"
	Recipient  = "recipient"
	Sender     = "sender"
)
