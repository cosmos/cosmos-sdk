package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC nft_transfer events
const (
	AttributeKeyReceiver = "receiver"
)

// IBC nft_transfer events vars
var (
	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
