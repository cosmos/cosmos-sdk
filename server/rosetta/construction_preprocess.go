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

	_, addr, _, err := OperationsToSdkMsg(operations)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidAddress, err.Error())
	}

	memo, ok := r.Metadata["memo"]
	if !ok {
		memo = ""
	}

	var res = &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			OptionAddress: addr,
			OptionGas:     200000,
			OptionMemo:    memo,
		},
	}
	return res, nil
}
