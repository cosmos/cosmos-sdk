package keeper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/armon/go-metrics"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) v1.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ v1.MsgServer = msgServer{}

func (k msgServer) SubmitProposal(goCtx context.Context, msg *v1.MsgSubmitProposal) (*v1.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposalMsgs, err := msg.GetMsgs()
	if err != nil {
		return nil, err
	}

	proposal, err := k.Keeper.SubmitProposal(ctx, proposalMsgs, msg.Metadata)
	if err != nil {
		return nil, err
	}

	bytes, err := proposal.Marshal()
	if err != nil {
		return nil, err
	}

	// ref: https://github.com/cosmos/cosmos-sdk/issues/9683
	ctx.GasMeter().ConsumeGas(
		3*storetypes.KVGasConfig().WriteCostPerByte*uint64(len(bytes)),
		"submit proposal",
	)

	defer telemetry.IncrCounter(1, types.ModuleName, "proposal")

	proposer, _ := sdk.AccAddressFromBech32(msg.GetProposer())
	votingStarted, err := k.Keeper.AddDeposit(ctx, proposal.Id, proposer, msg.GetInitialDeposit())
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetProposer()),
		),
	)

	if votingStarted {
		submitEvent := sdk.NewEvent(types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyVotingPeriodStart, fmt.Sprintf("%d", proposal.Id)),
		)

		ctx.EventManager().EmitEvent(submitEvent)
	}

	return &v1.MsgSubmitProposalResponse{
		ProposalId: proposal.Id,
	}, nil
}

func (k msgServer) ExecLegacyContent(goCtx context.Context, msg *v1.MsgExecLegacyContent) (*v1.MsgExecLegacyContentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAcct := k.GetGovernanceAccount(ctx).GetAddress().String()
	if govAcct != msg.Authority {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSigner, "expected %s got %s", govAcct, msg.Authority)
	}

	content, err := v1.LegacyContentFromMessage(msg)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidProposalContent, "%+v", err)
	}

	// Ensure that the content has a respective handler
	if !k.Keeper.legacyRouter.HasRoute(content.ProposalRoute()) {
		return nil, sdkerrors.Wrap(types.ErrNoProposalHandlerExists, content.ProposalRoute())
	}

	handler := k.Keeper.legacyRouter.GetRoute(content.ProposalRoute())
	if err := handler(ctx, content); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidProposalContent, "failed to run legacy handler %s, %+v", content.ProposalRoute(), err)
	}

	return &v1.MsgExecLegacyContentResponse{}, nil
}

func (k msgServer) Vote(goCtx context.Context, msg *v1.MsgVote) (*v1.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accAddr, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, err
	}
	err = k.Keeper.AddVote(ctx, msg.ProposalId, accAddr, v1.NewNonSplitVoteOption(msg.Option), msg.Metadata)
	if err != nil {
		return nil, err
	}

	defer telemetry.IncrCounterWithLabels(
		[]string{types.ModuleName, "vote"},
		1,
		[]metrics.Label{
			telemetry.NewLabel("proposal_id", strconv.Itoa(int(msg.ProposalId))),
		},
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter),
		),
	)

	return &v1.MsgVoteResponse{}, nil
}

func (k msgServer) VoteWeighted(goCtx context.Context, msg *v1.MsgVoteWeighted) (*v1.MsgVoteWeightedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accAddr, accErr := sdk.AccAddressFromBech32(msg.Voter)
	if accErr != nil {
		return nil, accErr
	}
	err := k.Keeper.AddVote(ctx, msg.ProposalId, accAddr, msg.Options, msg.Metadata)
	if err != nil {
		return nil, err
	}

	defer telemetry.IncrCounterWithLabels(
		[]string{types.ModuleName, "vote"},
		1,
		[]metrics.Label{
			telemetry.NewLabel("proposal_id", strconv.Itoa(int(msg.ProposalId))),
		},
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter),
		),
	)

	return &v1.MsgVoteWeightedResponse{}, nil
}

func (k msgServer) Deposit(goCtx context.Context, msg *v1.MsgDeposit) (*v1.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accAddr, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}
	votingStarted, err := k.Keeper.AddDeposit(ctx, msg.ProposalId, accAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	defer telemetry.IncrCounterWithLabels(
		[]string{types.ModuleName, "deposit"},
		1,
		[]metrics.Label{
			telemetry.NewLabel("proposal_id", strconv.Itoa(int(msg.ProposalId))),
		},
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor),
		),
	)

	if votingStarted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposalDeposit,
				sdk.NewAttribute(types.AttributeKeyVotingPeriodStart, fmt.Sprintf("%d", msg.ProposalId)),
			),
		)
	}

	return &v1.MsgDepositResponse{}, nil
}

type legacyMsgServer struct {
	govAcct string
	server  v1.MsgServer
}

// NewLegacyMsgServerImpl returns an implementation of the v1beta1 legacy MsgServer interface. It wraps around
// the current MsgServer
func NewLegacyMsgServerImpl(govAcct string, v1Server v1.MsgServer) v1beta1.MsgServer {
	return &legacyMsgServer{govAcct: govAcct, server: v1Server}
}

var _ v1beta1.MsgServer = legacyMsgServer{}

func (k legacyMsgServer) SubmitProposal(goCtx context.Context, msg *v1beta1.MsgSubmitProposal) (*v1beta1.MsgSubmitProposalResponse, error) {
	contentMsg, err := v1.NewLegacyContent(msg.GetContent(), k.govAcct)
	if err != nil {
		return nil, fmt.Errorf("error converting legacy content into proposal message: %w", err)
	}

	proposal, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{contentMsg},
		msg.InitialDeposit,
		msg.Proposer,
		"",
	)
	if err != nil {
		return nil, err
	}

	resp, err := k.server.SubmitProposal(goCtx, proposal)
	if err != nil {
		return nil, err
	}

	return &v1beta1.MsgSubmitProposalResponse{ProposalId: resp.ProposalId}, nil
}

func (k legacyMsgServer) Vote(goCtx context.Context, msg *v1beta1.MsgVote) (*v1beta1.MsgVoteResponse, error) {
	_, err := k.server.Vote(goCtx, &v1.MsgVote{
		ProposalId: msg.ProposalId,
		Voter:      msg.Voter,
		Option:     v1.VoteOption(msg.Option),
	})
	if err != nil {
		return nil, err
	}
	return &v1beta1.MsgVoteResponse{}, nil
}

func (k legacyMsgServer) VoteWeighted(goCtx context.Context, msg *v1beta1.MsgVoteWeighted) (*v1beta1.MsgVoteWeightedResponse, error) {
	opts := make([]*v1.WeightedVoteOption, len(msg.Options))
	for idx, opt := range msg.Options {
		opts[idx] = &v1.WeightedVoteOption{
			Option: v1.VoteOption(opt.Option),
			Weight: opt.Weight.String(),
		}
	}

	_, err := k.server.VoteWeighted(goCtx, &v1.MsgVoteWeighted{
		ProposalId: msg.ProposalId,
		Voter:      msg.Voter,
		Options:    opts,
	})
	if err != nil {
		return nil, err
	}
	return &v1beta1.MsgVoteWeightedResponse{}, nil
}

func (k legacyMsgServer) Deposit(goCtx context.Context, msg *v1beta1.MsgDeposit) (*v1beta1.MsgDepositResponse, error) {
	_, err := k.server.Deposit(goCtx, &v1.MsgDeposit{
		ProposalId: msg.ProposalId,
		Depositor:  msg.Depositor,
		Amount:     msg.Amount,
	})
	if err != nil {
		return nil, err
	}
	return &v1beta1.MsgDepositResponse{}, nil
}
