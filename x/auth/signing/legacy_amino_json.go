package signing

import (
	"fmt"

	v1beta12 "cosmossdk.io/api/cosmos/base/v1beta1"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/auth/migrations/legacytx"
	"cosmossdk.io/x/tx/decode"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

const aminoNonCriticalFieldsError = "protobuf transaction contains unknown non-critical fields. This is a transaction malleability issue and SIGN_MODE_LEGACY_AMINO_JSON cannot be used."

var _ SignModeHandler = signModeLegacyAminoJSONHandler{}

// signModeLegacyAminoJSONHandler defines the SIGN_MODE_LEGACY_AMINO_JSON
// SignModeHandler.
type signModeLegacyAminoJSONHandler struct {
	decoder *decode.Decoder
}

// NewSignModeLegacyAminoJSONHandler returns a new signModeLegacyAminoJSONHandler.
// Note: The public constructor is only used for testing.
// Deprecated: Please use x/tx/signing/aminojson instead.
func NewSignModeLegacyAminoJSONHandler(decoder *decode.Decoder) SignModeHandler {
	return signModeLegacyAminoJSONHandler{
		decoder: decoder,
	}
}

// Deprecated: Please use x/tx/signing/aminojson instead.
func (s signModeLegacyAminoJSONHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

// Deprecated: Please use x/tx/signing/aminojson instead.
func (s signModeLegacyAminoJSONHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

// Deprecated: Please use x/tx/signing/aminojson instead.
func (s signModeLegacyAminoJSONHandler) GetSignBytes(mode signingtypes.SignMode, data SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, mode)
	}

	protoTx, err := s.decoder.Decode(tx.Bytes())
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if protoTx.TxBodyHasUnknownNonCriticals {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, aminoNonCriticalFieldsError)
	}

	fees, err := getFees(protoTx.Tx.AuthInfo.Fee.Amount)
	if err != nil {
		return nil, err
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
		data.ChainID, data.AccountNumber, data.Sequence, protoTx.Tx.Body.TimeoutHeight,
		legacytx.StdFee{
			Amount:  fees,
			Gas:     protoTx.Tx.AuthInfo.Fee.GasLimit,
			Payer:   protoTx.Tx.AuthInfo.Fee.Payer,
			Granter: protoTx.Tx.AuthInfo.Fee.Granter,
		},
		tx.GetMsgs(), protoTx.Tx.Body.Memo,
	), nil
}

func getFees(fees []*v1beta12.Coin) (sdk.Coins, error) {
	f := make(sdk.Coins, len(fees))
	for i, fee := range fees {
		amtInt, ok := math.NewIntFromString(fee.Amount)
		if !ok {
			return nil, fmt.Errorf("invalid fee coin amount at index %d: %s", i, fee.Amount)
		}
		if err := sdk.ValidateDenom(fee.Denom); err != nil {
			return nil, fmt.Errorf("invalid fee coin denom at index %d: %w", i, err)
		}
		f[i] = sdk.Coin{
			Denom:  fee.Denom,
			Amount: amtInt,
		}
	}
	return f, nil
}
