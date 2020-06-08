package signing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

type SigningData struct {
	ChainID         string
	AccountNumber   uint64
	AccountSequence uint64
}

type SignModeHandler interface {
	DefaultMode() txtypes.SignMode
	Modes() []txtypes.SignMode
	GetSignBytes(mode txtypes.SignMode, data SigningData, tx sdk.Tx) ([]byte, error)
}
