package keeper

import (
	"context"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	ak AccountKeeper
}

// NewMsgServerImpl returns an implementation of the x/auth MsgServer interface.
func NewMsgServerImpl(ak AccountKeeper) types.MsgServer {
	return &msgServer{
		ak: ak,
	}
}

func (ms msgServer) AsyncMultiMsgExec(goCtx context.Context, msg *types.MsgAsyncMultiMsgExec) (*types.MsgAsyncMultiMsgExecResponse, error) {
	if msg.Signer == "" {
		return nil, errors.New("empty signer address string is not allowed")
	}

	signer, err := ms.ak.AddressCodec().StringToBytes(msg.Signer)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid signer address: %s", err)
	}

	if len(msg.Msgs) == 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("messages cannot be empty")
	}

	msgs, err := msg.GetMessages()
	if err != nil {
		return nil, err
	}

	if err := validateMsgs(msgs); err != nil {
		return nil, err
	}

	results, err := ms.ak.AsyncMsgsExec(goCtx, signer, msgs)
	if err != nil {
		return nil, err
	}

	return &types.MsgAsyncMultiMsgExecResponse{
		Error:   []string{},
		Results: results,
	}, nil
}

func validateMsgs(msgs []sdk.Msg) error {
	for i, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(err, "msg %d", i)
		}
	}

	return nil
}

func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.ak.authority != msg.Authority {
		return nil, fmt.Errorf(
			"expected authority account as only signer for proposal message; invalid authority; expected %s, got %s",
			ms.ak.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.ak.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
