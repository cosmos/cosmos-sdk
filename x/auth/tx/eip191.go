package tx

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const EIP191MessagePrefix = "\x19Ethereum Signed Message:\n"

const eip191NonCriticalFieldsError = "protobuf transaction contains unknown non-critical fields. This is a transaction malleability issue and SIGN_MODE_EIP_191 cannot be used."

var _ signing.SignModeHandler = signModeEIP191Handler{}

// signModeEIP191Handler defines the SIGN_MODE_EIP191
// SignModeHandler.
type signModeEIP191Handler struct{}

func (s signModeEIP191Handler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_EIP_191
}

func (s signModeEIP191Handler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_EIP_191}
}

func (s signModeEIP191Handler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_EIP_191 {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_EIP_191, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	if protoTx.txBodyHasUnknownNonCriticals {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, eip191NonCriticalFieldsError)
	}

	body := protoTx.tx.Body

	if len(body.ExtensionOptions) != 0 || len(body.NonCriticalExtensionOptions) != 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "SignMode_SIGN_MODE_EIP_191 does not support protobuf extension options.")
	}

	aminoJSONBz := legacytx.StdSignBytes(
		data.ChainID, data.AccountNumber, data.Sequence, protoTx.GetTimeoutHeight(),
		legacytx.StdFee{Amount: protoTx.GetFee(), Gas: protoTx.GetGas()},
		tx.GetMsgs(), protoTx.GetMemo(),
	)

	bz := append(
		[]byte(EIP191MessagePrefix),
		[]byte(strconv.Itoa(len(aminoJSONBz)))...,
	)

	bz = append(bz, aminoJSONBz...)

	return bz, nil
}
