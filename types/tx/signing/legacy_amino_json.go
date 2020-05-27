package signing

import (
	types "github.com/cosmos/cosmos-sdk/types/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type LegacyAminoJSONHandler struct{}

var _ SignModeHandler = LegacyAminoJSONHandler{}

func (LegacyAminoJSONHandler) Mode() types.SignMode {
	return types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

func (LegacyAminoJSONHandler) GetSignBytes(data SigningData, tx types.ProtoTx) ([]byte, error) {
	return auth.StdSignBytes(
		data.ChainID, data.AccountNumber, data.AccountSequence, auth.StdFee{Amount: tx.GetFee(), Gas: tx.GetGas()}, tx.GetMsgs(), tx.GetBody().Memo,
	), nil
}
