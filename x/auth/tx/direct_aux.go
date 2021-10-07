package tx

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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

	if data.Address == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "got empty address in SIGN_MODE_LEGACY_AMINO_JSON handler")
	}

	var pubKey *codectypes.Any
	pubKeys, err := protoTx.GetPubKeys()
	if err != nil {
		return nil, err
	}
	for i, pk := range pubKeys {
		if data.Address.Equals(sdk.AccAddress(pk.Address())) {
			pubKey = protoTx.tx.AuthInfo.SignerInfos[i].PublicKey
		}
	}
	if pubKey == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("got empty pubKey in SIGN_MODE_DIRECT_AUX handler")
	}

	signDocDirectAux := types.SignDocDirectAux{
		BodyBytes:     protoTx.getBodyBytes(),
		ChainId:       data.ChainID,
		AccountNumber: data.AccountNumber,
		Sequence:      data.Sequence,
		Tip:           protoTx.tx.AuthInfo.Tip,
		PublicKey:     pubKey,
	}

	return signDocDirectAux.Marshal()
}
