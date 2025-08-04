package keeper

import (
	"context"
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/evidence/types"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types2.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types2.MsgServer = msgServer{}

// SubmitEvidence implements the MsgServer.SubmitEvidence method.
func (ms msgServer) SubmitEvidence(goCtx context.Context, msg *types2.MsgSubmitEvidence) (*types2.MsgSubmitEvidenceResponse, error) {
	if _, err := ms.addressCodec.StringToBytes(msg.Submitter); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	evidence := msg.GetEvidence()
	if evidence == nil {
		return nil, errors.Wrap(types2.ErrInvalidEvidence, "missing evidence")
	}

	if err := evidence.ValidateBasic(); err != nil {
		return nil, errors.Wrapf(types2.ErrInvalidEvidence, "failed basic validation: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.Keeper.SubmitEvidence(ctx, evidence); err != nil {
		return nil, err
	}

	return &types2.MsgSubmitEvidenceResponse{
		Hash: evidence.Hash(),
	}, nil
}
