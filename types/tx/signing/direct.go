package signing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type DirectModeHandler struct{}

func (h DirectModeHandler) DefaultMode() types.SignMode {
	return types.SignMode_SIGN_MODE_DIRECT
}

var _ types.SignModeHandler = DirectModeHandler{}

func (DirectModeHandler) Modes() []types.SignMode {
	return []types.SignMode{types.SignMode_SIGN_MODE_DIRECT}
}

func (DirectModeHandler) GetSignBytes(mode types.SignMode, data types.SigningData, tx sdk.Tx) ([]byte, error) {
	protoTx, ok := tx.(types.ProtoTx)
	if !ok {
		return nil, fmt.Errorf("can only get direct sign bytes for a ProtoTx, got %T", tx)
	}
	bodyBz := protoTx.GetBodyBytes()
	authInfoBz := protoTx.GetAuthInfoBytes()
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
