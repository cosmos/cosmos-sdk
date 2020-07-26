package types

import (
	"fmt"

	sdk "github.com/KiraCore/cosmos-sdk/types"
	signingtypes "github.com/KiraCore/cosmos-sdk/types/tx/signing"
	"github.com/KiraCore/cosmos-sdk/x/auth/signing"
)

// LegacyAminoJSONHandler is a SignModeHandler that handles SIGN_MODE_LEGACY_AMINO_JSON
type LegacyAminoJSONHandler struct{}

var _ signing.SignModeHandler = LegacyAminoJSONHandler{}

// DefaultMode implements SignModeHandler.DefaultMode
func (h LegacyAminoJSONHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

// Modes implements SignModeHandler.Modes
func (LegacyAminoJSONHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

// DefaultMode implements SignModeHandler.GetSignBytes
func (LegacyAminoJSONHandler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, mode)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, fmt.Errorf("expected FeeTx, got %T", tx)
	}

	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return nil, fmt.Errorf("expected TxWithMemo, got %T", tx)
	}

	return StdSignBytes(
		data.ChainID, data.AccountNumber, data.AccountSequence, StdFee{Amount: feeTx.GetFee(), Gas: feeTx.GetGas()}, tx.GetMsgs(), memoTx.GetMemo(),
	), nil
}
