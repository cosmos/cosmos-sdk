package rosetta

import (
	"context"
	"encoding/hex"

	"github.com/tendermint/tendermint/crypto"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
)

func (l launchpad) ConstructionPayloads(ctx context.Context, req *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	if len(req.Operations) > 3 {
		return nil, ErrInvalidOperation
	}

	msgs, signAddr, fee, err := OperationsToSdkMsg(req.Operations)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidOperation, err.Error())
	}

	metadata, err := GetMetadataFromPayloadReq(req)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidRequest, err.Error())
	}

	tx := auth.NewStdTx(msgs, auth.StdFee{
		Amount: fee,
		Gas:    metadata.Gas,
	}, nil, metadata.Memo)
	signBytes := auth.StdSignBytes(
		metadata.ChainID, metadata.AccountNumber, metadata.Sequence, tx.Fee, tx.Msgs, tx.Memo,
	)
	txBytes, err := l.cdc.MarshalJSON(tx)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidRequest, err.Error())
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(txBytes),
		Payloads: []*types.SigningPayload{
			{
				AccountIdentifier: &types.AccountIdentifier{
					Address: signAddr,
				},
				Bytes:         crypto.Sha256(signBytes),
				SignatureType: "ecdsa",
			},
		},
	}, nil
}
