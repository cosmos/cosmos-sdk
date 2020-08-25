package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// stdTxSignModeHandler is a SignModeHandler that handles SIGN_MODE_LEGACY_AMINO_JSON
type stdTxSignModeHandler struct{}

func NewStdTxSignModeHandler() signing.SignModeHandler {
	return &stdTxSignModeHandler{}
}

var _ signing.SignModeHandler = stdTxSignModeHandler{}

// DefaultMode implements SignModeHandler.DefaultMode
func (h stdTxSignModeHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

// Modes implements SignModeHandler.Modes
func (stdTxSignModeHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

// DefaultMode implements SignModeHandler.GetSignBytes
func (stdTxSignModeHandler) GetSignBytes(mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, mode)
	}

	stdTx, ok := tx.(StdTx)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", StdTx{}, tx)
	}

	return StdSignBytes(
		data.ChainID, data.AccountNumber, data.Sequence, stdTx.GetTimeoutHeight(), StdFee{Amount: stdTx.GetFee(), Gas: stdTx.GetGas()}, tx.GetMsgs(), stdTx.GetMemo(),
	), nil
}
