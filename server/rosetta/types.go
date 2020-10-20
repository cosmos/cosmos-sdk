package rosetta

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	StatusReverted = "Reverted"
	StatusSuccess  = "Success"

	OperationTransfer = "Transfer"

	OptionAddress = "address"
	OptionGas     = "gas"
)

// TransferTxData represents a Tx that sends value.
type TransferTxData struct {
	From   types.AccAddress
	To     types.AccAddress
	Amount types.Coin
}

var Codec = simapp.MakeCodec()
