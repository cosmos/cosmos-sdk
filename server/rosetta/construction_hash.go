package rosetta

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func (l launchpad) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	//bz, err := hex.DecodeString(req.SignedTransaction)
	//if err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidTransaction, "error decoding tx")
	//}
	//
	//var stdTx authlegacy.StdTx
	//err = l.cdc.UnmarshalBinaryBare(bz, &stdTx)
	//if err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidTransaction, "invalid tx")
	//}
	//
	//txBytes, err := l.cdc.MarshalBinaryLengthPrefixed(stdTx)
	//if err != nil {
	//	return nil, rosetta.WrapError(ErrInvalidTransaction, "invalid tx")
	//}
	//
	//hash := sha256.Sum256(txBytes)
	//bzHash := hash[:]
	//
	//hashString := hex.EncodeToString(bzHash)
	//
	//return &types.TransactionIdentifierResponse{
	//	TransactionIdentifier: &types.TransactionIdentifier{
	//		Hash: strings.ToUpper(hashString),
	//	},
	//}, nil
	return nil, nil
}
