package rosetta

import (
	"context"
	"encoding/hex"

	"github.com/tendermint/tendermint/crypto"

	"github.com/coinbase/rosetta-sdk-go/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
)

func (l launchpad) ConstructionPayloads(ctx context.Context, req *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	// We only support for now Transfer type of operation.
	if len(req.Operations) != 2 {
		return nil, ErrInvalidOperation
	}

	if req.Operations[0].Type != OperationMsgSend || req.Operations[1].Type != OperationMsgSend {
		return nil, rosetta.WrapError(ErrInvalidOperation, "the operations are not Transfer")
	}

	transferData, err := getTransferTxDataFromOperations(req.Operations)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidOperation, err.Error())
	}

	msg := bank.NewMsgSend(transferData.From, transferData.To, cosmostypes.NewCoins(transferData.Amount))
	if err = msg.ValidateBasic(); err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, err.Error())
	}

	metadata, err := GetMetadataFromPayloadReq(req)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidRequest, err.Error())
	}

	tx := auth.NewStdTx([]cosmostypes.Msg{msg}, auth.StdFee{
		Gas: metadata.Gas,
	}, nil, "TODO memo") // TODO fees and memo.
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
				Address:       transferData.From.String(),
				Bytes:         crypto.Sha256(signBytes),
				SignatureType: "ecdsa",
			},
		},
	}, nil
}
