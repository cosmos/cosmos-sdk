package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

// SubmitProposal implements the MsgServer.SubmitProposal method.
func (k msgServer) SubmitProposal(goCtx context.Context, msg *v1.MsgSubmitProposal) (*v1.MsgSubmitProposalResponse, error) {
	if msg.Title == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("proposal title cannot be empty")
	}
	if msg.Summary == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("proposal summary cannot be empty")
	}

	proposer, err := k.authKeeper.AddressCodec().StringToBytes(msg.GetProposer())
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid proposer address: %s", err)
	}

	// check that either metadata or Msgs length is non nil.
	if len(msg.Messages) == 0 && len(msg.Metadata) == 0 {
		return nil, govtypes.ErrNoProposalMsgs.Wrap("either metadata or Msgs length must be non-nil")
	}

	// verify that if present, the metadata title and summary equals the proposal title and summary
	if len(msg.Metadata) != 0 {
		proposalMetadata := govtypes.ProposalMetadata{}
		if err := json.Unmarshal([]byte(msg.Metadata), &proposalMetadata); err == nil {
			if proposalMetadata.Title != msg.Title {
				return nil, govtypes.ErrInvalidProposalContent.Wrapf("metadata title '%s' must equal proposal title '%s'", proposalMetadata.Title, msg.Title)
			}

			if proposalMetadata.Summary != msg.Summary {
				return nil, govtypes.ErrInvalidProposalContent.Wrapf("metadata summary '%s' must equal proposal summary '%s'", proposalMetadata.Summary, msg.Summary)
			}
		}

		// if we can't unmarshal the metadata, this means the client didn't use the recommended metadata format
		// nothing can be done here, and this is still a valid case, so we ignore the error
	}

	proposalMsgs, err := msg.GetMsgs()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	initialDeposit := msg.GetInitialDeposit()

	params, err := k.Params.Get(goCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get governance parameters: %w", err)
	}

	if err := k.validateInitialDeposit(ctx, params, initialDeposit); err != nil {
		return nil, err
	}

	if err := k.validateDepositDenom(ctx, params, initialDeposit); err != nil {
		return nil, err
	}

	proposal, err := k.Keeper.SubmitProposal(ctx, proposalMsgs, msg.Metadata, msg.Title, msg.Summary, proposer)
	if err != nil {
		return nil, err
	}

	bytes, err := proposal.Marshal()
	if err != nil {
		return nil, err
	}

	// ref: https://github.com/cosmos/cosmos-sdk/issues/9683
	ctx.GasMeter().ConsumeGas(
		3*ctx.KVGasConfig().WriteCostPerByte*uint64(len(bytes)),
		"submit proposal",
	)

	// skip min deposit ratio check since for proposal submissions the initial deposit is the threshold
	// to check against.
	votingStarted, err := k.Keeper.AddDeposit(ctx, proposal.Id, proposer, msg.GetInitialDeposit(), true)
	if err != nil {
		return nil, err
	}

	if votingStarted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(govtypes.EventTypeSubmitProposal,
				sdk.NewAttribute(govtypes.AttributeKeyVotingPeriodStart, fmt.Sprintf("%d", proposal.Id)),
			),
		)
	}

	return &v1.MsgSubmitProposalResponse{
		ProposalId: proposal.Id,
	}, nil
}

// CancelProposal implements the MsgServer.CancelProposal method.
func (k msgServer) CancelProposal(goCtx context.Context, msg *v1.MsgCancelProposal) (*v1.MsgCancelProposalResponse, error) {
	_, err := k.authKeeper.AddressCodec().StringToBytes(msg.Proposer)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid proposer address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.CancelProposal(ctx, msg.ProposalId, msg.Proposer); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			govtypes.EventTypeCancelProposal,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Proposer),
			sdk.NewAttribute(govtypes.AttributeKeyProposalID, fmt.Sprint(msg.ProposalId)),
		),
	)

	return &v1.MsgCancelProposalResponse{
		ProposalId:     msg.ProposalId,
		CanceledTime:   ctx.BlockTime(),
		CanceledHeight: uint64(ctx.BlockHeight()),
	}, nil
}

// ExecLegacyContent implements the MsgServer.ExecLegacyContent method.
func (k msgServer) ExecLegacyContent(goCtx context.Context, msg *v1.MsgExecLegacyContent) (*v1.MsgExecLegacyContentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAcct := k.GetGovernanceAccount(ctx).GetAddress().String()
	if govAcct != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("expected %s got %s", govAcct, msg.Authority)
	}

	content, err := v1.LegacyContentFromMessage(msg)
	if err != nil {
		return nil, govtypes.ErrInvalidProposalContent.Wrapf("%+v", err)
	}

	// Ensure that the content has a respective handler
	if !k.Keeper.legacyRouter.HasRoute(content.ProposalRoute()) {
		return nil, govtypes.ErrNoProposalHandlerExists.Wrap(content.ProposalRoute())
	}

	handler := k.Keeper.legacyRouter.GetRoute(content.ProposalRoute())
	if err := handler(ctx, content); err != nil {
		return nil, govtypes.ErrInvalidProposalContent.Wrapf("failed to run legacy handler %s, %+v", content.ProposalRoute(), err)
	}

	return &v1.MsgExecLegacyContentResponse{}, nil
}

// Vote implements the MsgServer.Vote method.
func (k msgServer) Vote(goCtx context.Context, msg *v1.MsgVote) (*v1.MsgVoteResponse, error) {
	accAddr, err := k.authKeeper.AddressCodec().StringToBytes(msg.Voter)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid voter address: %s", err)
	}

	if !v1.ValidVoteOption(msg.Option) {
		return nil, govtypes.ErrInvalidVote.Wrap(msg.Option.String())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err = k.Keeper.AddVote(ctx, msg.ProposalId, accAddr, v1.NewNonSplitVoteOption(msg.Option), msg.Metadata)
	if err != nil {
		return nil, err
	}

	return &v1.MsgVoteResponse{}, nil
}

// VoteWeighted implements the MsgServer.VoteWeighted method.
func (k msgServer) VoteWeighted(goCtx context.Context, msg *v1.MsgVoteWeighted) (*v1.MsgVoteWeightedResponse, error) {
	accAddr, accErr := k.authKeeper.AddressCodec().StringToBytes(msg.Voter)
	if accErr != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid voter address: %s", accErr)
	}

	if len(msg.Options) == 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(v1.WeightedVoteOptions(msg.Options).String())
	}

	totalWeight := math.LegacyNewDec(0)
	usedOptions := make(map[v1.VoteOption]bool)
	for _, option := range msg.Options {
		if !option.IsValid() {
			return nil, govtypes.ErrInvalidVote.Wrap(option.String())
		}
		weight, err := math.LegacyNewDecFromStr(option.Weight)
		if err != nil {
			return nil, govtypes.ErrInvalidVote.Wrapf("invalid weight: %s", err)
		}
		totalWeight = totalWeight.Add(weight)
		if usedOptions[option.Option] {
			return nil, govtypes.ErrInvalidVote.Wrap("duplicated vote option")
		}
		usedOptions[option.Option] = true
	}

	if totalWeight.GT(math.LegacyNewDec(1)) {
		return nil, govtypes.ErrInvalidVote.Wrap("total weight overflow 1.00")
	}

	if totalWeight.LT(math.LegacyNewDec(1)) {
		return nil, govtypes.ErrInvalidVote.Wrap("total weight lower than 1.00")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.Keeper.AddVote(ctx, msg.ProposalId, accAddr, msg.Options, msg.Metadata)
	if err != nil {
		return nil, err
	}

	return &v1.MsgVoteWeightedResponse{}, nil
}

// Deposit implements the MsgServer.Deposit method.
func (k msgServer) Deposit(goCtx context.Context, msg *v1.MsgDeposit) (*v1.MsgDepositResponse, error) {
	accAddr, err := k.authKeeper.AddressCodec().StringToBytes(msg.Depositor)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid depositor address: %s", err)
	}

	if err := validateDeposit(msg.Amount); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	votingStarted, err := k.Keeper.AddDeposit(ctx, msg.ProposalId, accAddr, msg.Amount, false)
	if err != nil {
		return nil, err
	}

	if votingStarted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				govtypes.EventTypeProposalDeposit,
				sdk.NewAttribute(govtypes.AttributeKeyVotingPeriodStart, fmt.Sprintf("%d", msg.ProposalId)),
			),
		)
	}

	return &v1.MsgDepositResponse{}, nil
}

// UpdateParams implements the MsgServer.UpdateParams method.
func (k msgServer) UpdateParams(goCtx context.Context, msg *v1.MsgUpdateParams) (*v1.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	// before params change, trigger an update of the last min deposit
	ctx := sdk.UnwrapSDKContext(goCtx)
	blockTime := ctx.BlockTime()
	minDeposit := k.GetMinDeposit(ctx)
	newMinDeposit := v1.GetNewMinDeposit(msg.Params.MinDepositThrottler.FloorValue, minDeposit, math.LegacyOneDec())

	if !minDeposit.Equal(newMinDeposit) {
		err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
			Value: newMinDeposit,
			Time:  &blockTime,
		})
		if err != nil {
			return nil, err
		}
	}

	minInitialDeposit := k.GetMinInitialDeposit(ctx)
	newMinInitialDeposit := v1.GetNewMinDeposit(msg.Params.MinInitialDepositThrottler.FloorValue, minInitialDeposit, math.LegacyOneDec())

	if !minInitialDeposit.Equal(newMinInitialDeposit) {
		err := k.LastMinInitialDeposit.Set(goCtx, v1.LastMinDeposit{
			Value: newMinInitialDeposit,
			Time:  &blockTime,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.Params.Set(goCtx, msg.Params); err != nil {
		return nil, err
	}

	return &v1.MsgUpdateParamsResponse{}, nil
}

// ProposeLaw implements the MsgServer.ProposeLaw method.
func (k msgServer) ProposeLaw(_ context.Context, msg *v1.MsgProposeLaw) (*v1.MsgProposeLawResponse, error) {
	if k.authority != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}
	// only a no-op for now
	return &v1.MsgProposeLawResponse{}, nil
}

// ProposeConstitutionAmendment implements the MsgServer.ProposeConstitutionAmendment method.
func (k msgServer) ProposeConstitutionAmendment(goCtx context.Context, msg *v1.MsgProposeConstitutionAmendment) (*v1.MsgProposeConstitutionAmendmentResponse, error) {
	if k.authority != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}
	if msg.Amendment == "" {
		return nil, govtypes.ErrInvalidProposalMsg.Wrap("amendment cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	constitution, err := k.ApplyConstitutionAmendment(ctx, msg.Amendment)
	if err != nil {
		return nil, govtypes.ErrInvalidProposalMsg.Wrap(err.Error())
	}

	if err := k.Constitution.Set(goCtx, constitution); err != nil {
		return nil, err
	}

	return &v1.MsgProposeConstitutionAmendmentResponse{}, nil
}

func (k msgServer) CreateGovernor(goCtx context.Context, msg *v1.MsgCreateGovernor) (*v1.MsgCreateGovernorResponse, error) {
	// Ensure the governor does not already exist
	addr := sdk.MustAccAddressFromBech32(msg.GetAddress())
	govAddr := govtypes.GovernorAddress(addr.Bytes())
	if _, err := k.Governors.Get(goCtx, govAddr); err == nil {
		return nil, govtypes.ErrGovernorExists
	}

	// Ensure the governor has a valid description
	if _, err := msg.GetDescription().EnsureLength(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create the governor
	governor, err := v1.NewGovernor(govAddr.String(), msg.GetDescription(), ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	// validate min self-delegation
	params, err := k.Params.Get(goCtx)
	if err != nil {
		return nil, err
	}
	minSelfDelegation, _ := math.NewIntFromString(params.MinGovernorSelfDelegation)
	bondedTokens, err := k.getGovernorBondedTokens(ctx, govAddr)
	if err != nil {
		return nil, err
	}
	if bondedTokens.LT(minSelfDelegation) {
		return nil, govtypes.ErrInsufficientGovernorDelegation.Wrapf("minimum self-delegation required: %s, total bonded tokens: %s", minSelfDelegation, bondedTokens)
	}

	err = k.Governors.Set(goCtx, governor.GetAddress(), governor)
	if err != nil {
		return nil, err
	}

	// a base account automatically creates a governance delegation to itself
	// if a delegation to another governor already exists, undelegate first
	_, err = k.GovernanceDelegations.Get(goCtx, addr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if errors.Is(err, collections.ErrNotFound) {
		err = k.DelegateToGovernor(ctx, addr, govAddr)
		if err != nil {
			return nil, err
		}
	} else {
		err = k.RedelegateToGovernor(ctx, addr, govAddr)
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			govtypes.EventTypeCreateGovernor,
			sdk.NewAttribute(govtypes.AttributeKeyGovernor, govAddr.String()),
		),
	})

	return &v1.MsgCreateGovernorResponse{}, nil
}

func (k msgServer) EditGovernor(goCtx context.Context, msg *v1.MsgEditGovernor) (*v1.MsgEditGovernorResponse, error) {
	// Ensure the governor exists
	addr := sdk.MustAccAddressFromBech32(msg.GetAddress())
	govAddr := govtypes.GovernorAddress(addr.Bytes())
	governor, err := k.Governors.Get(goCtx, govAddr)
	if err != nil {
		return nil, govtypes.ErrGovernorNotFound
	}

	// Ensure the governor has a valid description
	if _, err := msg.GetDescription().EnsureLength(); err != nil {
		return nil, err
	}

	// Update the governor
	governor.Description = msg.GetDescription()
	err = k.Governors.Set(goCtx, governor.GetAddress(), governor)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			govtypes.EventTypeEditGovernor,
			sdk.NewAttribute(govtypes.AttributeKeyGovernor, govAddr.String()),
		),
	})

	return &v1.MsgEditGovernorResponse{}, nil
}

func (k msgServer) UpdateGovernorStatus(goCtx context.Context, msg *v1.MsgUpdateGovernorStatus) (*v1.MsgUpdateGovernorStatusResponse, error) {
	// Ensure the governor exists
	addr := sdk.MustAccAddressFromBech32(msg.GetAddress())
	govAddr := govtypes.GovernorAddress(addr.Bytes())
	governor, err := k.Governors.Get(goCtx, govAddr)
	if err != nil {
		return nil, govtypes.ErrGovernorNotFound
	}

	if !msg.GetStatus().IsValid() {
		return nil, govtypes.ErrInvalidGovernorStatus
	}

	// Ensure the governor is not already in the desired status
	if governor.Status == msg.GetStatus() {
		return nil, govtypes.ErrGovernorStatusEqual
	}

	// Ensure the governor has been in the current status for the required period
	params, err := k.Params.Get(goCtx)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	governorStatusChangePeriod := *params.GovernorStatusChangePeriod
	changeTime := ctx.BlockTime()
	if governor.LastStatusChangeTime.Add(governorStatusChangePeriod).After(changeTime) {
		return nil, govtypes.ErrGovernorStatusChangePeriod.Wrapf("last status change time: %s, need to wait until: %s", governor.LastStatusChangeTime, governor.LastStatusChangeTime.Add(governorStatusChangePeriod))
	}

	// Update the governor status
	governor.Status = msg.GetStatus()
	governor.LastStatusChangeTime = &changeTime
	// prevent a change to active if min self-delegation is not met
	if governor.IsActive() {
		if !k.ValidateGovernorMinSelfDelegation(ctx, governor) {
			return nil, govtypes.ErrInsufficientGovernorDelegation.Wrap("cannot set status to active: minimum self-delegation not met")
		}
	}

	err = k.Governors.Set(goCtx, governor.GetAddress(), governor)
	if err != nil {
		return nil, err
	}
	status := govtypes.AttributeValueStatusInactive
	// if status changes to active, create governance self-delegation
	// in case it didn't exist
	if governor.IsActive() {
		delegation, err := k.GovernanceDelegations.Get(goCtx, addr)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			panic(err)
		}
		if errors.Is(err, collections.ErrNotFound) {
			err := k.DelegateToGovernor(ctx, addr, govAddr)
			if err != nil {
				return nil, err
			}
		}
		if delegation.GovernorAddress != govAddr.String() {
			err := k.RedelegateToGovernor(ctx, addr, govAddr)
			if err != nil {
				return nil, err
			}
		}
		status = govtypes.AttributeValueStatusActive
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			govtypes.EventTypeGovernorChangeStatus,
			sdk.NewAttribute(govtypes.AttributeKeyGovernor, govAddr.String()),
			sdk.NewAttribute(govtypes.AttributeKeyStatus, status),
		),
	})

	return &v1.MsgUpdateGovernorStatusResponse{}, nil
}

func (k msgServer) DelegateGovernor(goCtx context.Context, msg *v1.MsgDelegateGovernor) (*v1.MsgDelegateGovernorResponse, error) {
	delAddr := sdk.MustAccAddressFromBech32(msg.GetDelegatorAddress())
	govAddr := govtypes.MustGovernorAddressFromBech32(msg.GetGovernorAddress())

	// Ensure the delegator is not already an active governor, as they cannot delegate
	if g, err := k.Governors.Get(goCtx, govtypes.GovernorAddress(delAddr.Bytes())); err == nil && g.IsActive() {
		return nil, govtypes.ErrDelegatorIsGovernor
	}

	// Ensure the delegation is not already present
	gd, err := k.GovernanceDelegations.Get(goCtx, delAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if err == nil && govAddr.Equals(govtypes.MustGovernorAddressFromBech32(gd.GovernorAddress)) {
		return nil, govtypes.ErrGovernanceDelegationExists
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// redelegate if a delegation to another governor already exists
	if err == nil {
		err := k.RedelegateToGovernor(ctx, delAddr, govAddr)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				govtypes.EventTypeRedelegate,
				sdk.NewAttribute(govtypes.AttributeKeySrcGovernor, gd.GovernorAddress),
				sdk.NewAttribute(govtypes.AttributeKeyDstGovernor, msg.GetGovernorAddress()),
				sdk.NewAttribute(govtypes.AttributeKeyDelegator, msg.GetDelegatorAddress()),
			),
		})
	} else {
		// Create the delegation
		err := k.DelegateToGovernor(ctx, delAddr, govAddr)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				govtypes.EventTypeDelegate,
				sdk.NewAttribute(govtypes.AttributeKeyDstGovernor, msg.GetGovernorAddress()),
				sdk.NewAttribute(govtypes.AttributeKeyDelegator, msg.GetDelegatorAddress()),
			),
		})
	}

	return &v1.MsgDelegateGovernorResponse{}, nil
}

func (k msgServer) UndelegateGovernor(goCtx context.Context, msg *v1.MsgUndelegateGovernor) (*v1.MsgUndelegateGovernorResponse, error) {
	delAddr := sdk.MustAccAddressFromBech32(msg.GetDelegatorAddress())

	// Ensure the delegation exists
	delegation, err := k.GovernanceDelegations.Get(goCtx, delAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if errors.Is(err, collections.ErrNotFound) {
		return nil, govtypes.ErrGovernanceDelegationNotFound
	}

	// if the delegator is also a governor, check if governor is active
	// if so, undelegation is not allowed. A status change to inactive is required first.
	delGovAddr := govtypes.GovernorAddress(delAddr.Bytes())
	delGovernor, err := k.Governors.Get(goCtx, delGovAddr)
	if err == nil && delGovernor.IsActive() {
		// if the delegation is not to self the state is inconsistent
		if delegation.GovernorAddress != delGovAddr.String() {
			panic("inconsistent state: active governor has a governance delegation to another governor")
		}
		return nil, govtypes.ErrDelegatorIsGovernor
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Remove the delegation
	err = k.UndelegateFromGovernor(ctx, delAddr)
	if err != nil {
		return nil, err
	}

	// if governor is inactive and does not have any delegations left, remove governor
	governor, err := k.Governors.Get(goCtx, govtypes.MustGovernorAddressFromBech32(delegation.GovernorAddress))
	if err != nil {
		panic("inconsistent state: governance delegation to non-existing governor")
	}
	if !governor.IsActive() {
		var delegations []*v1.GovernanceDelegation
		k.GovernanceDelegationsByGovernor.Walk(goCtx, collections.NewPrefixedPairRange[govtypes.GovernorAddress, sdk.AccAddress](governor.GetAddress()), func(_ collections.Pair[govtypes.GovernorAddress, sdk.AccAddress], value v1.GovernanceDelegation) (stop bool, err error) {
			delegations = append(delegations, &value)
			return false, nil
		})
		if len(delegations) == 0 {
			k.Governors.Remove(goCtx, governor.GetAddress())
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			govtypes.EventTypeUndelegate,
			sdk.NewAttribute(govtypes.AttributeKeySrcGovernor, delegation.GovernorAddress),
			sdk.NewAttribute(govtypes.AttributeKeyDelegator, msg.GetDelegatorAddress()),
		),
	})
	return &v1.MsgUndelegateGovernorResponse{}, nil
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
	content := msg.GetContent()
	if content == nil {
		return nil, govtypes.ErrInvalidProposalContent.Wrap("missing content")
	}
	if !v1beta1.IsValidProposalType(content.ProposalType()) {
		return nil, govtypes.ErrInvalidProposalType.Wrap(content.ProposalType())
	}
	if err := content.ValidateBasic(); err != nil {
		return nil, err
	}

	contentMsg, err := v1.NewLegacyContent(msg.GetContent(), k.govAcct)
	if err != nil {
		return nil, fmt.Errorf("error converting legacy content into proposal message: %w", err)
	}

	proposal, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{contentMsg},
		msg.InitialDeposit,
		msg.Proposer,
		"",
		msg.GetContent().GetTitle(),
		msg.GetContent().GetDescription(),
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

// validateDeposit validates the deposit amount, do not use for initial deposit.
func validateDeposit(amount sdk.Coins) error {
	if !amount.IsValid() || !amount.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap(amount.String())
	}

	return nil
}
