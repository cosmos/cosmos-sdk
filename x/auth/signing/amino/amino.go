package amino

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type LegacyAminoJSONHandler struct{}

var _ signing.SignModeHandler = LegacyAminoJSONHandler{}

func (h LegacyAminoJSONHandler) DefaultMode() types.SignMode {
	return types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

func (LegacyAminoJSONHandler) Modes() []types.SignMode {
	return []types.SignMode{types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

func (LegacyAminoJSONHandler) GetSignBytes(_ types.SignMode, data signing.SigningData, tx sdk.Tx) ([]byte, error) {
	feeTx, ok := tx.(ante.FeeTx)
	if !ok {
		return nil, fmt.Errorf("expected FeeTx, got %T", tx)
	}

	memoTx, ok := tx.(ante.TxWithMemo)
	if !ok {
		return nil, fmt.Errorf("expected TxWithMemo, got %T", tx)
	}

	return authtypes.StdSignBytes(
		data.ChainID, data.AccountNumber, data.AccountSequence, authtypes.StdFee{Amount: feeTx.GetFee(), Gas: feeTx.GetGas()}, tx.GetMsgs(), memoTx.GetMemo(),
	), nil
}
