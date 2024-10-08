package multisig

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/defaults/multisig/v1"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	MembersPrefix   = collections.NewPrefix(0)
	SequencePrefix  = collections.NewPrefix(1)
	ConfigPrefix    = collections.NewPrefix(2)
	ProposalsPrefix = collections.NewPrefix(3)
	VotesPrefix     = collections.NewPrefix(4)
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*Account)(nil)
)

type Account struct {
	Members  collections.Map[[]byte, uint64]
	Sequence collections.Sequence
	Config   collections.Item[v1.Config]

	addrCodec     address.Codec
	headerService header.Service
	eventService  event.Service

	Proposals collections.Map[uint64, v1.Proposal]
	Votes     collections.Map[collections.Pair[uint64, []byte], int32] // key: proposalID + voter address
}

// NewAccount returns a new multisig account creator function.
func NewAccount(deps accountstd.Dependencies) (*Account, error) {
	return &Account{
		Members:       collections.NewMap(deps.SchemaBuilder, MembersPrefix, "members", collections.BytesKey, collections.Uint64Value),
		Sequence:      collections.NewSequence(deps.SchemaBuilder, SequencePrefix, "sequence"),
		Config:        collections.NewItem(deps.SchemaBuilder, ConfigPrefix, "config", codec.CollValue[v1.Config](deps.LegacyStateCodec)),
		Proposals:     collections.NewMap(deps.SchemaBuilder, ProposalsPrefix, "proposals", collections.Uint64Key, codec.CollValue[v1.Proposal](deps.LegacyStateCodec)),
		Votes:         collections.NewMap(deps.SchemaBuilder, VotesPrefix, "votes", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey), collections.Int32Value),
		addrCodec:     deps.AddressCodec,
		headerService: deps.Environment.HeaderService,
		eventService:  deps.Environment.EventService,
	}, nil
}

// Init initializes the multisig account with the given configuration and members.
func (a *Account) Init(ctx context.Context, msg *v1.MsgInit) (*v1.MsgInitResponse, error) {
	if msg.Config == nil {
		return nil, errors.New("config must be specified")
	}

	if len(msg.Members) == 0 {
		return nil, errors.New("members must be specified")
	}

	// set members
	totalWeight := uint64(0)
	membersMap := map[string]struct{}{} // to check for duplicates
	for i := range msg.Members {
		if _, ok := membersMap[msg.Members[i].Address]; ok {
			return nil, errors.New("duplicate member address found")
		}

		membersMap[msg.Members[i].Address] = struct{}{}

		addrBz, err := a.addrCodec.StringToBytes(msg.Members[i].Address)
		if err != nil {
			return nil, err
		}

		if msg.Members[i].Weight == 0 {
			return nil, errors.New("member weight must be greater than zero")
		}

		if err := a.Members.Set(ctx, addrBz, msg.Members[i].Weight); err != nil {
			return nil, err
		}

		totalWeight, err = safeAdd(totalWeight, msg.Members[i].Weight)
		if err != nil {
			return nil, err
		}
	}

	if err := validateConfig(*msg.Config, totalWeight); err != nil {
		return nil, err
	}

	if err := a.Config.Set(ctx, *msg.Config); err != nil {
		return nil, err
	}

	return &v1.MsgInitResponse{}, nil
}

// Vote casts a vote on a proposal. The sender must be a member of the multisig and the proposal must be in the voting period.
func (a Account) Vote(ctx context.Context, msg *v1.MsgVote) (*v1.MsgVoteResponse, error) {
	if msg.Vote <= v1.VoteOption_VOTE_OPTION_UNSPECIFIED {
		return nil, errors.New("vote must be specified")
	}

	cfg, err := a.Config.Get(ctx)
	if err != nil {
		return nil, err
	}

	sender := accountstd.Sender(ctx)

	// check if the voter is a member
	_, err = a.Members.Get(ctx, sender)
	if err != nil {
		return nil, err
	}

	// check if the proposal exists
	prop, err := a.Proposals.Get(ctx, msg.ProposalId)
	if err != nil {
		return nil, err
	}

	// check if the voting period has ended
	if a.headerService.HeaderInfo(ctx).Time.Unix() > prop.VotingPeriodEnd || prop.Status != v1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		return nil, errors.New("voting period has ended")
	}

	// check if the voter has already voted
	_, err = a.Votes.Get(ctx, collections.Join(msg.ProposalId, sender))
	if err == nil && !cfg.Revote {
		return nil, errors.New("voter has already voted, can't change its vote per config")
	}
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}

	addr, err := a.addrCodec.BytesToString(sender)
	if err != nil {
		return nil, err
	}

	if err = a.eventService.EventManager(ctx).EmitKV("vote",
		event.NewAttribute("proposal_id", fmt.Sprint(msg.ProposalId)),
		event.NewAttribute("voter", addr),
		event.NewAttribute("vote", msg.Vote.String()),
	); err != nil {
		return nil, err
	}

	return &v1.MsgVoteResponse{}, a.Votes.Set(ctx, collections.Join(msg.ProposalId, sender), int32(msg.Vote))
}

// CreateProposal creates a new proposal. If the proposal contains a voting
// period it will be used, otherwise the default voting period will be used.
func (a Account) CreateProposal(ctx context.Context, msg *v1.MsgCreateProposal) (*v1.MsgCreateProposalResponse, error) {
	// check if the sender is a member
	_, err := a.Members.Get(ctx, accountstd.Sender(ctx))
	if err != nil {
		return nil, err
	}

	seq, err := a.Sequence.Next(ctx)
	if err != nil {
		return nil, err
	}

	// check if the proposal already exists
	_, err = a.Proposals.Get(ctx, seq)
	if err == nil {
		return nil, errors.New("proposal already exists")
	}

	config, err := a.Config.Get(ctx)
	if err != nil {
		return nil, err
	}

	// create the proposal
	proposal := v1.Proposal{
		Title:    msg.Proposal.Title,
		Summary:  msg.Proposal.Summary,
		Messages: msg.Proposal.Messages,
		Status:   v1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
	}

	// set the voting period, if not specified, use the default
	if msg.Proposal.VotingPeriodEnd != 0 {
		proposal.VotingPeriodEnd = msg.Proposal.VotingPeriodEnd
	} else {
		proposal.VotingPeriodEnd = a.headerService.HeaderInfo(ctx).Time.Add(time.Second * time.Duration(config.VotingPeriod)).Unix()
	}

	if err = a.Proposals.Set(ctx, seq, proposal); err != nil {
		return nil, err
	}

	addr, err := a.addrCodec.BytesToString(accountstd.Sender(ctx))
	if err != nil {
		return nil, err
	}

	if err = a.eventService.EventManager(ctx).EmitKV("proposal_created",
		event.NewAttribute("proposal_id", fmt.Sprint(seq)),
		event.NewAttribute("proposer", addr),
	); err != nil {
		return nil, err
	}

	return &v1.MsgCreateProposalResponse{ProposalId: seq}, nil
}

// deleteProposalAndVotes deletes a proposal and its votes, pruning the state.
func (a Account) deleteProposalAndVotes(ctx context.Context, proposalID uint64) error {
	// delete the proposal
	if err := a.Proposals.Remove(ctx, proposalID); err != nil {
		return err
	}

	// delete the votes
	rng := collections.NewPrefixedPairRange[uint64, []byte](proposalID)
	return a.Votes.Clear(ctx, rng)
}

// ExecuteProposal tallies the votes for a proposal and executes it if it passes. If early execution is enabled, it will
// ignore the voting period and tally the votes without deleting them if the proposal has not passed.
func (a Account) ExecuteProposal(ctx context.Context, msg *v1.MsgExecuteProposal) (*v1.MsgExecuteProposalResponse, error) {
	prop, err := a.Proposals.Get(ctx, msg.ProposalId)
	if err != nil {
		return nil, err
	}

	config, err := a.Config.Get(ctx)
	if err != nil {
		return nil, err
	}

	// check if voting period is still active and early execution is disabled
	votingPeriodEnded := a.headerService.HeaderInfo(ctx).Time.Unix() > prop.VotingPeriodEnd
	if !votingPeriodEnded && !config.EarlyExecution {
		return nil, errors.New("voting period has not ended yet, and early execution is not enabled")
	}

	// perform tally
	rng := collections.NewPrefixedPairRange[uint64, []byte](msg.ProposalId)
	yesVotes := uint64(0)
	noVotes := uint64(0)
	abstainVotes := uint64(0)
	err = a.Votes.Walk(ctx, rng, func(key collections.Pair[uint64, []byte], vote int32) (stop bool, err error) {
		weight, err := a.Members.Get(ctx, key.K2())
		if errors.Is(err, collections.ErrNotFound) {
			// edge case: if a member has been removed after voting, we should ignore their vote
			return false, nil
		} else if err != nil {
			return true, err
		}

		switch v1.VoteOption(vote) {
		case v1.VoteOption_VOTE_OPTION_YES:
			yesVotes += weight
		case v1.VoteOption_VOTE_OPTION_NO:
			noVotes += weight
		case v1.VoteOption_VOTE_OPTION_ABSTAIN:
			abstainVotes += weight
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	totalWeight := yesVotes + noVotes + abstainVotes

	var (
		rejectErr error
		execErr   error
	)

	resp := &v1.MsgExecuteProposalResponse{}

	if totalWeight < uint64(config.Quorum) {
		rejectErr = errors.New("quorum not reached")
		prop.Status = v1.ProposalStatus_PROPOSAL_STATUS_REJECTED
	} else if yesVotes < uint64(config.Threshold) {
		rejectErr = errors.New("threshold not reached")
		prop.Status = v1.ProposalStatus_PROPOSAL_STATUS_REJECTED
	} else {
		// we have quorum and threshold, execute the proposal
		prop.Status = v1.ProposalStatus_PROPOSAL_STATUS_PASSED
		resp.Responses, execErr = accountstd.ExecModuleAnys(ctx, prop.Messages) // do not return this error, just emit the event
	}

	// if early execution is enabled, we return early if the proposal has NOT passed
	if config.EarlyExecution && prop.Status != v1.ProposalStatus_PROPOSAL_STATUS_PASSED {
		return nil, errors.New("early execution attempted and proposal has not passed")
	}

	if err = a.deleteProposalAndVotes(ctx, msg.ProposalId); err != nil {
		return nil, err
	}

	if err = a.eventService.EventManager(ctx).EmitKV("proposal_tally",
		event.NewAttribute("proposal_id", fmt.Sprint(msg.ProposalId)),
		event.NewAttribute("yes_votes", fmt.Sprint(yesVotes)),
		event.NewAttribute("no_votes", fmt.Sprint(noVotes)),
		event.NewAttribute("abstain_votes", fmt.Sprint(abstainVotes)),
		event.NewAttribute("status", prop.Status.String()),
		event.NewAttribute("reject_err", fmt.Sprint(rejectErr)),
		event.NewAttribute("exec_err", fmt.Sprint(execErr)),
	); err != nil {
		return nil, err
	}

	err = a.Proposals.Set(ctx, msg.ProposalId, prop)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// QuerySequence returns the current sequence number, used for proposal IDs.
func (a Account) QuerySequence(ctx context.Context, _ *v1.QuerySequence) (*v1.QuerySequenceResponse, error) {
	seq, err := a.Sequence.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.QuerySequenceResponse{Sequence: seq}, nil
}

// QueryProposal returns a proposal for a given ID, if the proposal hasn't been pruned yet.
func (a Account) QueryProposal(ctx context.Context, q *v1.QueryProposal) (*v1.QueryProposalResponse, error) {
	proposal, err := a.Proposals.Get(ctx, q.ProposalId)
	if err != nil {
		return nil, err
	}
	return &v1.QueryProposalResponse{Proposal: &proposal}, nil
}

// QueryConfig returns the current multisig configuration.
func (a Account) QueryConfig(ctx context.Context, _ *v1.QueryConfig) (*v1.QueryConfigResponse, error) {
	cfg, err := a.Config.Get(ctx)
	if err != nil {
		return nil, err
	}

	members := []*v1.Member{}
	err = a.Members.Walk(ctx, nil, func(addr []byte, weight uint64) (stop bool, err error) {
		addrStr, err := a.addrCodec.BytesToString(addr)
		if err != nil {
			return true, err
		}
		members = append(members, &v1.Member{Address: addrStr, Weight: weight})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &v1.QueryConfigResponse{Config: &cfg, Members: members}, nil
}

// RegisterExecuteHandlers implements implementation.Account.
func (a *Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.Vote)
	accountstd.RegisterExecuteHandler(builder, a.CreateProposal)
	accountstd.RegisterExecuteHandler(builder, a.ExecuteProposal)
	accountstd.RegisterExecuteHandler(builder, a.UpdateConfig)
}

// RegisterInitHandler implements implementation.Account.
func (a *Account) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

// RegisterQueryHandlers implements implementation.Account.
func (a *Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QuerySequence)
	accountstd.RegisterQueryHandler(builder, a.QueryProposal)
	accountstd.RegisterQueryHandler(builder, a.QueryConfig)
}

func safeAdd(nums ...uint64) (uint64, error) {
	var sum uint64
	for _, num := range nums {
		if newSum := sum + num; newSum < sum {
			return 0, errors.New("overflow")
		} else {
			sum = newSum
		}
	}
	return sum, nil
}
