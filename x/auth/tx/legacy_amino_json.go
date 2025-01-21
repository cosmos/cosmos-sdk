package tx

import (
	"fmt"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const aminoNonCriticalFieldsError = "protobuf transaction contains unknown non-critical fields. This is a transaction malleability issue and SIGN_MODE_LEGACY_AMINO_JSON cannot be used."

var _ signing.SignModeHandler = signModeLegacyAminoJSONHandler{}

// signModeLegacyAminoJSONHandler defines the SIGN_MODE_LEGACY_AMINO_JSON
// SignModeHandler.
type signModeLegacyAminoJSONHandler struct{}

// NewSignModeLegacyAminoJSONHandler returns a new signModeLegacyAminoJSONHandler.
// Note: The public constructor is only used for testing.
// Deprecated: Please use x/tx/signing/aminojson instead.
func NewSignModeLegacyAminoJSONHandler() signing.SignModeHandler {
	return signModeLegacyAminoJSONHandler{}
}

// Deprecated: Please use x/tx/signing/aminojson instead.
func (s signModeLegacyAminoJSONHandler) DefaultMode() apisigning.SignMode {
	return apisigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

// Deprecated: Please use x/tx/signing/aminojson instead.
func (s signModeLegacyAminoJSONHandler) Modes() []apisigning.SignMode {
	return []apisigning.SignMode{apisigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

// Deprecated: Please use x/tx/signing/aminojson instead.
func (s signModeLegacyAminoJSONHandler) GetSignBytes(mode apisigning.SignMode, data txsigning.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != apisigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, mode)
	}

	protoTx, ok := tx.(*gogoTxWrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	if protoTx.TxBodyHasUnknownNonCriticals {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, aminoNonCriticalFieldsError)
	}

	body := protoTx.Tx.Body

	if len(body.ExtensionOptions) != 0 || len(body.NonCriticalExtensionOptions) != 0 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s does not support protobuf extension options", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	addr := data.Address
	if addr == "" {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "got empty address in %s handler", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	return legacytx.StdSignBytes(
		data.ChainID, data.AccountNumber, data.Sequence, protoTx.GetTimeoutHeight(),
		legacytx.StdFee{
			Amount:  protoTx.GetFee(),
			Gas:     protoTx.GetGas(),
			Payer:   protoTx.Tx.AuthInfo.Fee.Payer,
			Granter: protoTx.Tx.AuthInfo.Fee.Granter,
		},
		tx.GetMsgs(), protoTx.GetMemo(),
	), nil
}
