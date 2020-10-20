package rosetta

import (
	"context"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func (l launchpad) ConstructionPreprocess(ctx context.Context, r *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	operations := r.Operations
	if len(operations) < 1 {
		return nil, ErrInterpreting
	}

	txData, err := getTransferTxDataFromOperations(operations)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidAddress, err.Error())
	}
	if txData.From == nil {
		return nil, rosetta.WrapError(ErrInvalidAddress, err.Error())
	}

	var res = &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			OptionAddress: txData.From.String(),
			OptionGas:     200000,
		},
	}
	return res, nil
}
