package tx

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const aminoNonCriticalFieldsError = "protobuf transaction contains unknown non-critical fields. This is a transaction malleability issue and SIGN_MODE_LEGACY_AMINO_JSON cannot be used."

var _ signing.SignModeHandler = signModeLegacyAminoJSONHandler{}

// signModeLegacyAminoJSONHandler defines the SIGN_MODE_LEGACY_AMINO_JSON
// SignModeHandler.
type signModeLegacyAminoJSONHandler struct{}

func (s signModeLegacyAminoJSONHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

func (s signModeLegacyAminoJSONHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

func (s signModeLegacyAminoJSONHandler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, mode)
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
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s does not support protobuf extension options", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	addr := data.Address
	if addr == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "got empty address in %s handler", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	seq, err := getSequence(protoTx, addr)
	if err != nil {
		return nil, err
	}

	// We set a convention that if the tipper signs with LEGACY_AMINO_JSON, then
	// they sign over empty fees and 0 gas.
	var isTipper bool
	if addr == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "got empty address in SIGN_MODE_LEGACY_AMINO_JSON handler")
	}
	if tipTx, ok := tx.(typestx.TipTx); ok && tipTx.GetTip() != nil {
		isTipper = tipTx.GetTip().Tipper == addr.String()
	}

	if isTipper {
		return legacytx.StdSignBytes(
			data.ChainID, data.AccountNumber, data.Sequence, protoTx.GetTimeoutHeight(),
			// The tipper signs over 0 fee and 0 gas, no feepayer, no feegranter by convention.
			legacytx.StdFee{},
			tx.GetMsgs(), protoTx.GetMemo(),
		), nil
	}

	return legacytx.StdSignBytes(
		data.ChainID, data.AccountNumber, seq, protoTx.GetTimeoutHeight(),
		legacytx.StdFee{Amount: protoTx.GetFee(), Gas: protoTx.GetGas(), Payer: protoTx.FeePayer().String(), Granter: protoTx.FeeGranter().String()},
		tx.GetMsgs(), protoTx.GetMemo(),
	), nil
}

// getSequence retrieves the sequence of the given address from the protoTx's
// signer infos.
func getSequence(protoTx *wrapper, addr sdk.AccAddress) (uint64, error) {
	sigsV2, err := protoTx.GetSignaturesV2()
	if err != nil {
		return 0, err
	}
	for _, si := range sigsV2 {
		if addr.Equals(sdk.AccAddress(si.PubKey.Address())) {
			return si.Sequence, nil
		}
	}

	return 0, sdkerrors.ErrInvalidRequest.Wrapf("address %s not found in signer infos", addr)
}
