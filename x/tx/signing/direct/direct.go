package direct

import (
	"context"

	gogoproto "github.com/cosmos/gogoproto/proto"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/tx/signing"
)

var _ signing.SignModeHandler = SignModeHandler{}

// SignModeHandler is the SIGN_MODE_DIRECT implementation of signing.SignModeHandler.
type SignModeHandler struct{}

// Mode implements signing.SignModeHandler.Mode.
func (h SignModeHandler) Mode() signing.SignMode {
	return signing.SignMode_SIGN_MODE_DIRECT
}

// GetSignBytes implements signing.SignModeHandler.GetSignBytes.
// SignDoc has only scalar fields (no maps), so gogoproto.Marshal produces
// the same deterministic bytes as proto.MarshalOptions{Deterministic:true}.
func (SignModeHandler) GetSignBytes(_ context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {
	return gogoproto.Marshal(&txtypes.SignDoc{
		BodyBytes:     txData.BodyBytes,
		AuthInfoBytes: txData.AuthInfoBytes,
		ChainId:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
	})
}
