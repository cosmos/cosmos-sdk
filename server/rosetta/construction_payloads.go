package rosetta

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func (l launchpad) ConstructionPayloads(ctx context.Context, req *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	//// We only support for now Transfer type of operation.
	//if len(req.Operations) != 2 {
	//	return nil, ErrInvalidOperation
	//}
	//
	//if req.Operations[0].Type != OperationTransfer || req.Operations[1].Type != OperationTransfer {
	//	return nil, rosetta.WrapError(ErrInvalidOperation, "the operations are not Transfer")
	//}
	//
	//transferData, err := getTransferTxDataFromOperations(req.Operations)
	//if err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidOperation, err.Error())
	//}
	//
	//msg := bank.NewMsgSend(transferData.From, transferData.To, cosmostypes.NewCoins(transferData.Amount))
	//if err = msg.ValidateBasic(); err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidTransaction, err.Error())
	//}
	//
	//metadata, err := GetMetadataFromPayloadReq(req)
	//if err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidRequest, err.Error())
	//}
	//
	//tx := authlegacy.NewStdTx([]cosmostypes.Msg{msg}, authlegacy.StdFee{
	//	Gas: metadata.Gas,
	//}, nil, "TODO memo") // TODO fees and memo.
	//signBytes := authlegacy.StdSignBytes(
	//	metadata.ChainID, metadata.AccountNumber, metadata.Sequence, 9999999, tx.Fee, tx.Msgs, tx.Memo,
	//)
	//txBytes, err := l.cdc.MarshalBinaryBare(tx)
	//if err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidRequest, err.Error())
	//}
	//
	//return &types.ConstructionPayloadsResponse{
	//	UnsignedTransaction: hex.EncodeToString(txBytes),
	//	Payloads: []*types.SigningPayload{
	//		{
	//			AccountIdentifier: &types.AccountIdentifier{
	//				Address: transferData.From.String(),
	//			},
	//			Bytes:         crypto.Sha256(signBytes),
	//			SignatureType: "ecdsa",
	//		},
	//	},
	//}, nil
	return nil, nil
}
