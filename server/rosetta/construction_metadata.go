package rosetta

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
)

func (l launchpad) ConstructionMetadata(ctx context.Context, r *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	if l.properties.OfflineMode {
		return nil, ErrEndpointDisabledOfflineMode
	}

	if len(r.Options) == 0 {
		return nil, ErrInvalidRequest
	}

	addr, ok := r.Options[OptionAddress]
	if !ok {
		return nil, ErrInvalidAddress
	}
	addrString := addr.(string)
	accRes, err := l.cosmos.GetAuthAccount(ctx, addrString, int64(0))
	if err != nil {
		return nil, rosetta.WrapError(ErrInterpreting, err.Error())
	}

	gas, ok := r.Options[GasKey]
	if !ok {
		return nil, rosetta.WrapError(ErrInvalidAddress, "gas not set")
	}

	memo, ok := r.Options[OptionMemo]
	if !ok {
		return nil, ErrInvalidMemo
	}

	statusRes, err := l.tendermint.Status()
	if err != nil {
		return nil, rosetta.WrapError(ErrInterpreting, err.Error())
	}

	res := &types.ConstructionMetadataResponse{
		Metadata: map[string]interface{}{
			AccountNumberKey: accRes.Result.Value.AccountNumber,
			SequenceKey:      accRes.Result.Value.Sequence,
			ChainIDKey:       statusRes.NodeInfo.Network,
			GasKey:           gas,
			OptionMemo:       memo,
		},
	}

	return res, nil
}
