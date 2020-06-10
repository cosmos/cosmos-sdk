package amino

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// LegacyAminoJSONHandler is a SignModeHandler that handles SIGN_MODE_LEGACY_AMINO_JSON
type LegacyAminoJSONHandler struct{}

var _ signing.SignModeHandler = LegacyAminoJSONHandler{}

// DefaultMode implements SignModeHandler.DefaultMode
func (h LegacyAminoJSONHandler) DefaultMode() types.SignMode {
	return types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

// Modes implements SignModeHandler.Modes
func (LegacyAminoJSONHandler) Modes() []types.SignMode {
	return []types.SignMode{types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

// DefaultMode implements SignModeHandler.GetSignBytes
func (LegacyAminoJSONHandler) GetSignBytes(mode types.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error) {
	if mode != types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
		return nil, fmt.Errorf("expected %s, got %s", types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, mode)
	}

	feeTx, ok := tx.(ante.FeeTx)
	if !ok {
		return nil, fmt.Errorf("expected FeeTx, got %T", tx)
	}

	memoTx, ok := tx.(ante.TxWithMemo)
	if !ok {
		return nil, fmt.Errorf("expected TxWithMemo, got %T", tx)
	}

	return authtypes.StdSignBytes(
		data.ChainID, data.AccountNumber, data.AccountSequence, authtypes.StdFee{Amount: feeTx.GetFee(), Gas: feeTx.GetGas()}, tx.GetMsgs(), memoTx.GetMemo(), // nolint:staticcheck // SA1019: authtypes.StdFee is deprecated, will be removed once proto migration is completed
	), nil
}
