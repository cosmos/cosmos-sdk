package rosetta

import (
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	StatusReverted = "Reverted"
	StatusSuccess  = "Success"

	OperationMsgSend  = "send"
	OperationDelegate = "delegate"

	OptionAddress = "address"
	OptionGas     = "gas"
	OperationFee  = "fee"
	OptionMemo    = "memo"
)

// TransferTxData represents a Tx that sends value.
type TransferTxData struct {
	From   types.AccAddress
	To     types.AccAddress
	Amount types.Coin
}
