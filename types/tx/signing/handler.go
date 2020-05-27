package signing

import (
	"github.com/tendermint/tendermint/crypto"

	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type SigningData struct {
	ModeInfo        *types.ModeInfo_Single
	PublicKey       crypto.PubKey
	ChainID         string
	AccountNumber   uint64
	AccountSequence uint64
}

type SignModeHandler interface {
	Mode() types.SignMode
	GetSignBytes(data SigningData, tx types.ProtoTx) ([]byte, error)
}
