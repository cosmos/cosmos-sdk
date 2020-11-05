package rosetta

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// ConstructionParse implements the /construction/parse endpoint.
func (l launchpad) ConstructionParse(ctx context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	//rawTx, err := hex.DecodeString(request.Transaction)
	//if err != nil {
	//	return nil, ErrInvalidTransaction
	//}
	//
	//var stdTx authlegacy.StdTx
	//err = l.cdc.UnmarshalBinaryBare(rawTx, &stdTx)
	//if err != nil {
	//	return nil, ErrInvalidTransaction
	//}
	//
	//signers := make([]string, len(stdTx.Signatures))
	//for i, sig := range stdTx.Signatures {
	//	addr, err := cosmostypes.AccAddressFromHex(sig.PubKey.Address().String())
	//	if err != nil {
	//		return nil, ErrInvalidTransaction
	//	}
	//	signers[i] = addr.String()
	//}
	//
	//var accIdentifiers []*types.AccountIdentifier
	//for _, signer := range signers {
	//	accIdentifiers = append(accIdentifiers, &types.AccountIdentifier{
	//		Address: signer,
	//	})
	//}
	//return &types.ConstructionParseResponse{
	//	Operations:               toOperations(stdTx.Msgs, false, true),
	//	AccountIdentifierSigners: accIdentifiers,
	//}, nil
	return nil, nil
}
