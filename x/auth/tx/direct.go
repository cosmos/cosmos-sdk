package tx

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// signModeDirectHandler defines the SIGN_MODE_DIRECT SignModeHandler
type signModeDirectHandler struct{}

var _ signing.SignModeHandler = signModeDirectHandler{}

// DefaultMode implements SignModeHandler.DefaultMode
func (signModeDirectHandler) DefaultMode() signing.SignMode {
	return signing.SignMode_SIGN_MODE_DIRECT
}

// Modes implements SignModeHandler.Modes
func (signModeDirectHandler) Modes() []signing.SignMode {
	return []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (signModeDirectHandler) GetSignBytes(mode signing.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signing.SignMode_SIGN_MODE_DIRECT {
		return nil, fmt.Errorf("expected %s, got %s", signing.SignMode_SIGN_MODE_DIRECT, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	bodyBz := protoTx.getBodyBytes()
	authInfoBz := protoTx.getAuthInfoBytes()

	return DirectSignBytes(bodyBz, authInfoBz, data.ChainID, data.AccountNumber)
}

// DirectSignBytes returns the SIGN_MODE_DIRECT sign bytes for the provided TxBody bytes, AuthInfo bytes, chain ID,
// account number and sequence.
func DirectSignBytes(bodyBytes, authInfoBytes []byte, chainID string, accnum uint64) ([]byte, error) {
	signDoc := sdktx.SignDoc{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		ChainId:       chainID,
		AccountNumber: accnum,
	}
	return signDoc.Marshal()
}
