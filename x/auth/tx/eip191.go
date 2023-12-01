package tx

import (
	"context"
	"strconv"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
)

const EIP191MessagePrefix = "\x19Ethereum Signed Message:\n"

const eip191NonCriticalFieldsError = "protobuf transaction contains unknown non-critical fields. This is a transaction malleability issue and SIGN_MODE_EIP_191 cannot be used."

// SignModeEIP191Handler defines the SIGN_MODE_DIRECT SignModeHandler
type SignModeEIP191Handler struct {
	*aminojson.SignModeHandler
}

// NewSignModeEIP191Handler returns a new SignModeEIP191Handler.
func NewSignModeEIP191Handler(options aminojson.SignModeHandlerOptions) *SignModeEIP191Handler {
	return &SignModeEIP191Handler{
		SignModeHandler: aminojson.NewSignModeHandler(options),
	}
}

var _ signing.SignModeHandler = SignModeEIP191Handler{}

// Mode implements signing.SignModeHandler.Mode.
func (SignModeEIP191Handler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_EIP_191
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (h SignModeEIP191Handler) GetSignBytes(
	ctx context.Context, data signing.SignerData, txData signing.TxData,
) ([]byte, error) {
	aminoJSONBz, err := h.SignModeHandler.GetSignBytes(ctx, data, txData)
	if err != nil {
		return nil, err
	}

	return append(append(
		[]byte(EIP191MessagePrefix),
		[]byte(strconv.Itoa(len(aminoJSONBz)))...,
	), aminoJSONBz...), nil
}
