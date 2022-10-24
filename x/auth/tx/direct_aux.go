package tx

import (
	"fmt"

	codectypes "github.com/pointnetwork/cosmos-point-sdk/codec/types"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	sdkerrors "github.com/pointnetwork/cosmos-point-sdk/types/errors"
	types "github.com/pointnetwork/cosmos-point-sdk/types/tx"
	signingtypes "github.com/pointnetwork/cosmos-point-sdk/types/tx/signing"
	"github.com/pointnetwork/cosmos-point-sdk/x/auth/signing"
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

	pkAny, err := codectypes.NewAnyWithValue(data.PubKey)
	if err != nil {
		return nil, err
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
		Sequence:      data.Sequence,
		Tip:           protoTx.tx.AuthInfo.Tip,
		PublicKey:     pkAny,
	}

	return signDocDirectAux.Marshal()
}
