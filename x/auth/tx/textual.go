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

var _ signing.SignModeHandler = signModeTextualHandler{}

// signModeTextualHandler defines the SIGN_MODE_TEXTUAL SignModeHandler
type signModeTextualHandler struct{}

// DefaultMode implements SignModeHandler.DefaultMode
func (signModeTextualHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_TEXTUAL
}

// Modes implements SignModeHandler.Modes
func (signModeTextualHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_TEXTUAL}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (signModeTextualHandler) GetSignBytes(
	mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx,
) ([]byte, error) {

	if mode != signingtypes.SignMode_SIGN_MODE_TEXTUAL {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_TEXTUAL, mode)
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
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "got empty address in %s handler", signingtypes.SignMode_SIGN_MODE_TEXTUAL)
	}

	feePayer := protoTx.FeePayer().String()

	// Fee payer cannot use SIGN_MODE_TEXTUAL, because SIGN_MODE_TEXTUAL
	// does not sign over fees, which would create malleability issues.
	if feePayer == data.Address {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("fee payer %s cannot sign with %s", feePayer, signingtypes.SignMode_SIGN_MODE_TEXTUAL)
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
