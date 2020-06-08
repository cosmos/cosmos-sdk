package signing

import (
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SigningData struct {
	ChainID         string
	AccountNumber   uint64
	AccountSequence uint64
	PublicKey       crypto.PubKey
}

type SignModeHandler interface {
	DefaultMode() txtypes.SignMode
	Modes() []txtypes.SignMode
	GetSignBytes(mode txtypes.SignMode, data SigningData, tx sdk.Tx) ([]byte, error)
}
