package tx

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ signing.SignModeHandler = signModeDirectAuxHandler{}

// signModeDirectAuxHandler defines the SIGN_MODE_DIRECT_AUX SignModeHandler
type signModeDirectAuxHandler struct{}

// DefaultMode implements SignModeHandler.DefaultMode
func (signModeDirectAuxHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_DIRECT_AUX
}

// Modes implements SignModeHandler.Modes
func (signModeDirectAuxHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT_AUX}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (signModeDirectAuxHandler) GetSignBytes(
	mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx,
) ([]byte, error) {

	if mode != signingtypes.SignMode_SIGN_MODE_DIRECT_AUX {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_DIRECT_AUX, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	signerInfo := protoTx.tx.AuthInfo.SignerInfos[data.SignerIndex]
	if signerInfo == nil || signerInfo.PublicKey == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("got empty pubkey for signer #%d in %s handler", data.SignerIndex, signingtypes.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	addr := data.Address
	if addr == "" {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "got empty address in %s handler", signingtypes.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	feePayer := protoTx.FeePayer().String()

	// Fee payer cannot use SIGN_MODE_DIRECT_AUX, because SIGN_MODE_DIRECT_AUX
	// does not sign over fees, which would create malleability issues.
	if feePayer == data.Address {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("fee payer %s cannot sign with %s", feePayer, signingtypes.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	signDocDirectAux := types.SignDocDirectAux{
		BodyBytes:     protoTx.getBodyBytes(),
		ChainId:       data.ChainID,
		AccountNumber: data.AccountNumber,
		Sequence:      signerInfo.Sequence,
		Tip:           protoTx.tx.AuthInfo.Tip,
		PublicKey:     signerInfo.PublicKey,
	}

	return signDocDirectAux.Marshal()
}
