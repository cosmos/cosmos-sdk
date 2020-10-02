package keeper

import (
	"context"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	v1 "github.com/cosmos/cosmos-sdk/types/module/v1"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type msgServer struct {
	key v1.StoreKey
}

func NewMsgServer() types.MsgServer {
	return msgServer{}
}

func (m msgServer) SubmitProposal(ctx context.Context, _ *types.MsgSubmitProposal) (*types.MsgSubmitProposalResponse, error) {
	store := m.key.KVStore(ctx)
	store.Get([]byte("test"))
	bankClient := banktypes.NewMsgClient(m.key)
	_, err := bankClient.Send(ctx, &banktypes.MsgSend{})
	if err != nil {
		return nil, err
	}
	return &types.MsgSubmitProposalResponse{}, nil
}
