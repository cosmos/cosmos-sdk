package tx

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// signModeDirectAuxHandler defines the SIGN_MODE_DIRECT_AUX SignModeHandler
type signModeDirectAuxHandler struct{}

var _ signing.SignModeHandler = signModeDirectAuxHandler{}

// DefaultMode implements SignModeHandler.DefaultMode
func (signModeDirectAuxHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_DIRECT_AUX
}

// Modes implements SignModeHandler.Modes
func (signModeDirectAuxHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT_AUX}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (signModeDirectAuxHandler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_DIRECT_AUX {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_DIRECT_AUX, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	signDocDirectAux := types.SignDocDirectAux{
		BodyBytes:     protoTx.getBodyBytes(),
		ChainId:       data.ChainID,
		AccountNumber: data.AccountNumber,
		Sequence:      data.Sequence,
		Tip:           protoTx.tx.AuthInfo.Tip,
		// PublicKey: ,
	}
	return signDocDirectAux.Marshal()
	// return DirectAuxSignBytes(bodyBz, authInfoBz, data.ChainID, data.AccountNumber, data.Sequence)
}

// // DirectAuxSignBytes returns the SignMode_SIGN_MODE_DIRECT_AUX sign bytes for the provided TxBody bytes, AuthInfo bytes, chain ID,
// // account number and sequence.
// func DirectAuxSignBytes(bodyBytes, authInfoBytes []byte, chainID string, accnum, seq uint64) ([]byte, error) {
// 	signDocDirectAux := types.SignDocDirectAux{
// 		BodyBytes:     bodyBytes,
// 		ChainId:       chainID,
// 		AccountNumber: accnum,
// 		Sequence:      seq,
// 	}

// 	return signDocDirectAux.Marshal()
// }
