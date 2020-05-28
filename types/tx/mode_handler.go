package types

import (
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SigningData struct {
	ModeInfo        *ModeInfo_Single
	PublicKey       crypto.PubKey
	ChainID         string
	AccountNumber   uint64
	AccountSequence uint64
}

type SignModeHandler interface {
	Modes() []SignMode
	GetSignBytes(data SigningData, tx sdk.Tx) ([]byte, error)
}
