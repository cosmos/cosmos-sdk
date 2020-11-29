package rosetta

import (
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	StatusReverted = "Reverted"
	StatusSuccess  = "Success"

	OperationMsgSend = "cosmos-sdk/MsgSend"

	OptionAddress = "address"
	OptionGas     = "gas"
	OperationFee  = "fee"
)

// TransferTxData represents a Tx that sends value.
type TransferTxData struct {
	From   types.AccAddress
	To     types.AccAddress
	Amount types.Coin
}
