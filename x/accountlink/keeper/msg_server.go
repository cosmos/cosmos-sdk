package keeper

import (
	"context"

	"cosmossdk.io/x/accountlink/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) Register(ctx context.Context, msg *types.MsgRegister) (*types.MsgRegisterResponse, error) {
	err := msg.Validate()
	if err != nil {
		return nil, err
	}

	err = m.CheckCondition(ctx, types.Condition{
		Owner:    msg.Owner,
		Account:  msg.Account,
		Messages: msg.Messages,
	})
	if err != nil {
		return nil, err
	}

	if accExists := m.authKeeper.HasAccount(ctx, sdk.AccAddress(msg.Owner)); !accExists {
		return nil, types.ErrNonExistOwner
	}

	err = m.Keeper.SetAccountsByOwner(ctx, sdk.AccAddress(msg.Owner), msg.AccountType, msg.Account)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterResponse{}, nil
}
