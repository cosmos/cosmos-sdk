package tx

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ signing.SignModeHandler = signModeAminoAuxHandler{}

// signModeAminoAuxHandler defines the SIGN_MODE_AMINO_AUX SignModeHandler
type signModeAminoAuxHandler struct{}

// DefaultMode implements SignModeHandler.DefaultMode
func (signModeAminoAuxHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_AMINO_AUX
}

// Modes implements SignModeHandler.Modes
func (signModeAminoAuxHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_AMINO_AUX}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (signModeAminoAuxHandler) GetSignBytes(
	mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx,
) ([]byte, error) {

	if mode != signingtypes.SignMode_SIGN_MODE_AMINO_AUX {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_AMINO_AUX, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	if protoTx.txBodyHasUnknownNonCriticals {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, aminoNonCriticalFieldsError)
	}

	body := protoTx.tx.Body

	if len(body.ExtensionOptions) != 0 || len(body.NonCriticalExtensionOptions) != 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "SIGN_MODE_AMINO_AUX does not support protobuf extension options.")
	}

	return legacytx.StdSignAuxBytes(
		data.ChainID, data.AccountNumber, data.Sequence, protoTx.GetTimeoutHeight(),
		protoTx.tx.AuthInfo.Tip.Amount, tx.GetMsgs(), protoTx.GetMemo(),
	), nil
}
