package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/x/bank/v2/types"
)

type handlers struct {
	*Keeper
}

// NewHandlers creates a new bank/v2 handlers
func NewHandlers(k *Keeper) handlers {
	return handlers{k}
}

// UpdateParams updates the parameters of the bank/v2 module.
func (h handlers) MsgUpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	authorityBytes, err := h.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(h.authority, authorityBytes) {
		expectedAuthority, err := h.addressCodec.BytesToString(h.authority)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("invalid authority; expected %s, got %s", expectedAuthority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := h.params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// QueryParams queries the parameters of the bank/v2 module.
func (h handlers) QueryParams(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	params, err := h.params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}
