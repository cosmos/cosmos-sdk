package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/rekeying/types"
)

type msgServer struct {
	authkeeper.AccountKeeper
	rekeyingKeeper Keeper
}

// NewMsgServerImpl returns an implementation of the changepubkey MsgServer interface,
// wrapping the corresponding AccountKeeper.
func NewMsgServerImpl(k authkeeper.AccountKeeper, rekeyingKeeper Keeper) types.MsgServer {
	return &msgServer{k, rekeyingKeeper}
}

var _ types.MsgServer = msgServer{}

func (s msgServer) ChangePubKey(goCtx context.Context, msg *types.MsgChangePubKey) (*types.MsgChangePubKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := s.AccountKeeper

	accAddress, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}
	acc := ak.GetAccount(ctx, accAddress)
	if acc == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s does not exist", msg.Address)
	}

	pubKey, err := msg.GetPubKey(s.rekeyingKeeper.cdc)
	if err != nil {
		return nil, err
	}

	addressByPubkey := sdk.AccAddress(pubKey.Address())
	// check if pubKey already on the chain.
	if acc := ak.GetAccount(ctx, addressByPubkey); acc != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", addressByPubkey)
	}

	err = acc.SetPubKey(pubKey)
	if err != nil {
		return nil, err
	}
	ak.SetAccount(ctx, acc)

	return &types.MsgChangePubKeyResponse{}, nil
}
