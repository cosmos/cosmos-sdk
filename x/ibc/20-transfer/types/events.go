package types

import (
	"fmt"

	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
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
	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, transfer.SubModuleName)
)
