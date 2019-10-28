package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC transfer events
const (
	AttributeKeySender   = "sender"
	AttributeKeyReceiver = "receiver"
	AttributeKeyAmount   = "amount"
)

// IBC transfer events vars
var (
	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
