package signing

import (
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type DirectModeHandler struct{}

var _ SignModeHandler = DirectModeHandler{}

func (DirectModeHandler) Mode() types.SignMode {
	return types.SignMode_SIGN_MODE_DIRECT
}

func (DirectModeHandler) GetSignBytes(data SigningData, tx types.ProtoTx) ([]byte, error) {
	bodyBz := tx.GetBodyBytes()
	authInfoBz := tx.GetAuthInfoBytes()
	return DirectSignBytes(bodyBz, authInfoBz, data.ChainID, data.AccountNumber, data.AccountSequence)
}

func DirectSignBytes(bodyBz, authInfoBz []byte, chainID string, accnum, sequence uint64) ([]byte, error) {
	signDoc := SignDocRaw{
		BodyBytes:       bodyBz,
		AuthInfoBytes:   authInfoBz,
		ChainId:         chainID,
		AccountNumber:   accnum,
		AccountSequence: sequence,
	}
	return signDoc.Marshal()
}
