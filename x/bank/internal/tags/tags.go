package tags

import (
	"github.com/cosmos/cosmos-sdk/types"
)

// Tag keys and values
var (
	ActionUndelegateCoins = "undelegateCoins"
	ActionDelegateCoins   = "delegateCoins"
	TxCategory            = "bank"

	Action    = types.TagAction
	Category  = types.TagCategory
	Recipient = "recipient"
	Sender    = "sender"
)
