package signing

import (
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

type LegacyAminoJSONHandler struct{}

var _ SignModeHandler = LegacyAminoJSONHandler{}

func (LegacyAminoJSONHandler) Mode() types.SignMode {
	return types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

func (LegacyAminoJSONHandler) GetSignBytes(data SigningData, tx DecodedTx) ([]byte, error) {
	return auth.StdSignBytes(
		data.ChainID, data.AccountNumber, data.AccountSequence, auth.StdFee{Amount: tx.GetFee(), Gas: tx.GetGas()}, tx.Msgs, tx.Body.Memo,
	), nil
}
