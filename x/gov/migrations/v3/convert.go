package v3

import (
	"fmt"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// ConvertToLegacyProposal takes a new proposal and attempts to convert it to the
// legacy proposal format. This conversion is best effort. New proposal types that
// don't have a legacy message will return a "nil" content.
// Returns error when the amount of messages in `proposal` is different than one.
func ConvertToLegacyProposal(proposal v1.Proposal) (v1beta1.Proposal, error) {
	var err error
	legacyProposal := v1beta1.Proposal{
		ProposalId:   proposal.Id,
		Status:       v1beta1.ProposalStatus(proposal.Status),
		TotalDeposit: types.NewCoins(proposal.TotalDeposit...),
	}

	legacyProposal.FinalTallyResult, err = ConvertToLegacyTallyResult(proposal.FinalTallyResult)
	if err != nil {
		return v1beta1.Proposal{}, err
	}

	if proposal.VotingStartTime != nil {
		legacyProposal.VotingStartTime = *proposal.VotingStartTime
	}

	if proposal.VotingEndTime != nil {
		legacyProposal.VotingEndTime = *proposal.VotingEndTime
	}

	if proposal.SubmitTime != nil {
		legacyProposal.SubmitTime = *proposal.SubmitTime
	}

	if proposal.DepositEndTime != nil {
		legacyProposal.DepositEndTime = *proposal.DepositEndTime
	}

	msgs, err := proposal.GetMsgs()
	if err != nil {
		return v1beta1.Proposal{}, err
	}
	if len(msgs) != 1 {
		return v1beta1.Proposal{}, sdkerrors.ErrInvalidType.Wrap("can't convert a gov/v1 Proposal to gov/v1beta1 Proposal when amount of proposal messages not exactly one")
	}
	if legacyMsg, ok := msgs[0].(*v1.MsgExecLegacyContent); ok {
		// check that the content struct can be unmarshalled
		_, err := v1.LegacyContentFromMessage(legacyMsg)
		if err != nil {
			return v1beta1.Proposal{}, err
		}
		legacyProposal.Content = legacyMsg.Content
		return legacyProposal, nil
	}
	// hack to fill up the content with the first message
	// this is to support clients that have not yet (properly) use gov/v1 endpoints
	// https://github.com/cosmos/cosmos-sdk/issues/14334
	// VerifyBasic assures that we have at least one message.
	legacyProposal.Content, err = codectypes.NewAnyWithValue(msgs[0])

	return legacyProposal, err
}

func ConvertToLegacyTallyResult(tally *v1.TallyResult) (v1beta1.TallyResult, error) {
	yes, ok := math.NewIntFromString(tally.YesCount)
	if !ok {
		return v1beta1.TallyResult{}, fmt.Errorf("unable to convert yes tally string (%s) to int", tally.YesCount)
	}
	no, ok := math.NewIntFromString(tally.NoCount)
	if !ok {
		return v1beta1.TallyResult{}, fmt.Errorf("unable to convert no tally string (%s) to int", tally.NoCount)
	}

	abstain, ok := math.NewIntFromString(tally.AbstainCount)
	if !ok {
		return v1beta1.TallyResult{}, fmt.Errorf("unable to convert abstain tally string (%s) to int", tally.AbstainCount)
	}

	return v1beta1.TallyResult{
		Yes:        yes,
		No:         no,
		Abstain:    abstain,
		NoWithVeto: math.ZeroInt(),
	}, nil
}

func ConvertToLegacyVote(vote v1.Vote) (v1beta1.Vote, error) {
	options, err := ConvertToLegacyVoteOptions(vote.Options)
	if err != nil {
		return v1beta1.Vote{}, err
	}
	return v1beta1.Vote{
		ProposalId: vote.ProposalId,
		Voter:      vote.Voter,
		Options:    options,
	}, nil
}

func ConvertToLegacyVoteOptions(voteOptions []*v1.WeightedVoteOption) ([]v1beta1.WeightedVoteOption, error) {
	options := make([]v1beta1.WeightedVoteOption, len(voteOptions))
	for i, option := range voteOptions {
		weight, err := math.LegacyNewDecFromStr(option.Weight)
		if err != nil {
			return options, err
		}
		options[i] = v1beta1.WeightedVoteOption{
			Option: v1beta1.VoteOption(option.Option),
			Weight: weight,
		}
	}
	return options, nil
}

func ConvertToLegacyDeposit(deposit *v1.Deposit) v1beta1.Deposit {
	return v1beta1.Deposit{
		ProposalId: deposit.ProposalId,
		Depositor:  deposit.Depositor,
		Amount:     types.NewCoins(deposit.Amount...),
	}
}

func convertToNewProposal(oldProp v1beta1.Proposal) (v1.Proposal, error) {
	msg, err := v1.NewLegacyContent(oldProp.GetContent(), authtypes.NewModuleAddress("gov").String())
	if err != nil {
		return v1.Proposal{}, err
	}
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return v1.Proposal{}, err
	}

	return v1.Proposal{
		Id:       oldProp.ProposalId,
		Messages: []*codectypes.Any{msgAny},
		Status:   v1.ProposalStatus(oldProp.Status),
		FinalTallyResult: &v1.TallyResult{
			YesCount:     oldProp.FinalTallyResult.Yes.String(),
			NoCount:      oldProp.FinalTallyResult.No.String(),
			AbstainCount: oldProp.FinalTallyResult.Abstain.String(),
		},
		SubmitTime:      &oldProp.SubmitTime,
		DepositEndTime:  &oldProp.DepositEndTime,
		TotalDeposit:    oldProp.TotalDeposit,
		VotingStartTime: &oldProp.VotingStartTime,
		VotingEndTime:   &oldProp.VotingEndTime,
		Title:           oldProp.GetContent().GetTitle(),
		Summary:         oldProp.GetContent().GetDescription(),
	}, nil
}

func convertToNewProposals(oldProps v1beta1.Proposals) (v1.Proposals, error) {
	newProps := make([]*v1.Proposal, len(oldProps))
	for i, oldProp := range oldProps {
		p, err := convertToNewProposal(oldProp)
		if err != nil {
			return nil, err
		}

		newProps[i] = &p
	}

	return newProps, nil
}
