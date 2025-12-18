package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	v3 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v3"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var _ v1.QueryServer = queryServer{}

type queryServer struct{ k *Keeper }

func NewQueryServer(k *Keeper) v1.QueryServer {
	return queryServer{k: k}
}

func (q queryServer) Constitution(ctx context.Context, _ *v1.QueryConstitutionRequest) (*v1.QueryConstitutionResponse, error) {
	constitution, err := q.k.Constitution.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	return &v1.QueryConstitutionResponse{Constitution: constitution}, nil
}

// Proposal returns proposal details based on ProposalID
func (q queryServer) Proposal(ctx context.Context, req *v1.QueryProposalRequest) (*v1.QueryProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	proposal, err := q.k.Proposals.Get(ctx, req.ProposalId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "proposal %d doesn't exist", req.ProposalId)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryProposalResponse{Proposal: &proposal}, nil
}

// Proposals implements the Query/Proposals gRPC method
func (q queryServer) Proposals(ctx context.Context, req *v1.QueryProposalsRequest) (*v1.QueryProposalsResponse, error) {
	filteredProposals, pageRes, err := query.CollectionFilteredPaginate(ctx, q.k.Proposals, req.Pagination, func(_ uint64, p v1.Proposal) (include bool, err error) {
		matchVoter, matchDepositor, matchStatus := true, true, true

		// match status (if supplied/valid)
		if v1.ValidProposalStatus(req.ProposalStatus) {
			matchStatus = p.Status == req.ProposalStatus
		}

		// match voter address (if supplied)
		if len(req.Voter) > 0 {
			voter, err := q.k.authKeeper.AddressCodec().StringToBytes(req.Voter)
			if err != nil {
				return false, err
			}

			has, err := q.k.Votes.Has(ctx, collections.Join(p.Id, sdk.AccAddress(voter)))
			// if no error, vote found, matchVoter = true
			matchVoter = err == nil && has
		}

		// match depositor (if supplied)
		if len(req.Depositor) > 0 {
			depositor, err := q.k.authKeeper.AddressCodec().StringToBytes(req.Depositor)
			if err != nil {
				return false, err
			}
			has, err := q.k.Deposits.Has(ctx, collections.Join(p.Id, sdk.AccAddress(depositor)))
			// if no error, deposit found, matchDepositor = true
			matchDepositor = err == nil && has
		}

		// if all match, append to results
		if matchVoter && matchDepositor && matchStatus {
			return true, nil
		}
		// continue to next item, do not include because we're appending results above.
		return false, nil
	}, func(_ uint64, value v1.Proposal) (*v1.Proposal, error) {
		return &value, nil
	})

	if err != nil && !errors.Is(err, collections.ErrInvalidIterator) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryProposalsResponse{Proposals: filteredProposals, Pagination: pageRes}, nil
}

// Vote returns Voted information based on proposalID, voterAddr
func (q queryServer) Vote(ctx context.Context, req *v1.QueryVoteRequest) (*v1.QueryVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	if req.Voter == "" {
		return nil, status.Error(codes.InvalidArgument, "empty voter address")
	}

	voter, err := q.k.authKeeper.AddressCodec().StringToBytes(req.Voter)
	if err != nil {
		return nil, err
	}
	vote, err := q.k.Votes.Get(ctx, collections.Join(req.ProposalId, sdk.AccAddress(voter)))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.InvalidArgument,
				"voter: %v not found for proposal: %v", req.Voter, req.ProposalId)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryVoteResponse{Vote: &vote}, nil
}

// Votes returns single proposal's votes
func (q queryServer) Votes(ctx context.Context, req *v1.QueryVotesRequest) (*v1.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	votes, pageRes, err := query.CollectionPaginate(ctx, q.k.Votes, req.Pagination, func(_ collections.Pair[uint64, sdk.AccAddress], value v1.Vote) (vote *v1.Vote, err error) {
		return &value, nil
	}, query.WithCollectionPaginationPairPrefix[uint64, sdk.AccAddress](req.ProposalId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryVotesResponse{Votes: votes, Pagination: pageRes}, nil
}

// Params queries all params
func (q queryServer) Params(ctx context.Context, req *v1.QueryParamsRequest) (*v1.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// NOTE: feed deprecated parameters with dynamic values for backward compat
	params.MinDeposit = q.k.GetMinDeposit(ctx)
	params.Quorum = q.k.GetQuorum(ctx).String()
	params.ConstitutionAmendmentQuorum = q.k.GetConstitutionAmendmentQuorum(ctx).String()
	params.LawQuorum = q.k.GetLawQuorum(ctx).String()

	response := &v1.QueryParamsResponse{}

	//nolint:staticcheck // needed for legacy parameters
	switch req.ParamsType {
	case v1.ParamDeposit:
		depositParams := v1.NewDepositParams(params.MinDeposit, params.MaxDepositPeriod)
		response.DepositParams = &depositParams

	case v1.ParamVoting:
		votingParams := v1.NewVotingParams(params.VotingPeriod)
		response.VotingParams = &votingParams

	case v1.ParamTallying:
		tallyParams := v1.NewTallyParams(
			params.Quorum, params.Threshold,
			params.ConstitutionAmendmentQuorum, params.ConstitutionAmendmentThreshold,
			params.LawQuorum, params.LawThreshold,
		)
		response.TallyParams = &tallyParams
	default:
		if len(req.ParamsType) > 1 {
			return nil, status.Errorf(codes.InvalidArgument, "unknown params type: %s", req.ParamsType)
		}
	}
	response.Params = &params

	return response, nil
}

// Deposit queries single deposit information based on proposalID, depositAddr.
func (q queryServer) Deposit(ctx context.Context, req *v1.QueryDepositRequest) (*v1.QueryDepositResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	if req.Depositor == "" {
		return nil, status.Error(codes.InvalidArgument, "empty depositor address")
	}

	depositor, err := q.k.authKeeper.AddressCodec().StringToBytes(req.Depositor)
	if err != nil {
		return nil, err
	}
	deposit, err := q.k.Deposits.Get(ctx, collections.Join(req.ProposalId, sdk.AccAddress(depositor)))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &v1.QueryDepositResponse{Deposit: &deposit}, nil
}

// Deposits returns single proposal's all deposits
func (q queryServer) Deposits(ctx context.Context, req *v1.QueryDepositsRequest) (*v1.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	var deposits []*v1.Deposit
	deposits, pageRes, err := query.CollectionPaginate(ctx, q.k.Deposits, req.Pagination, func(_ collections.Pair[uint64, sdk.AccAddress], deposit v1.Deposit) (*v1.Deposit, error) {
		return &deposit, nil
	}, query.WithCollectionPaginationPairPrefix[uint64, sdk.AccAddress](req.ProposalId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}

// TallyResult queries the tally of a proposal vote
func (q queryServer) TallyResult(ctx context.Context, req *v1.QueryTallyResultRequest) (*v1.QueryTallyResultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	proposal, err := q.k.Proposals.Get(ctx, req.ProposalId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "proposal %d doesn't exist", req.ProposalId)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	var tallyResult v1.TallyResult

	switch {
	case proposal.Status == v1.StatusDepositPeriod:
		tallyResult = v1.EmptyTallyResult()

	case proposal.Status == v1.StatusPassed || proposal.Status == v1.StatusRejected || proposal.Status == v1.StatusFailed:
		tallyResult = *proposal.FinalTallyResult

	default:
		// proposal is in voting period
		var err error
		_, _, _, tallyResult, err = q.k.Tally(ctx, proposal)
		if err != nil {
			return nil, err
		}
	}

	return &v1.QueryTallyResultResponse{Tally: &tallyResult}, nil
}

// MinDeposit returns the minimum deposit currently required for a proposal to enter voting period
func (q queryServer) MinDeposit(c context.Context, _ *v1.QueryMinDepositRequest) (*v1.QueryMinDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minDeposit := q.k.GetMinDeposit(ctx)

	return &v1.QueryMinDepositResponse{MinDeposit: minDeposit}, nil
}

// MinInitialDeposit returns the minimum deposit required for a proposal to be submitted
func (q queryServer) MinInitialDeposit(c context.Context, _ *v1.QueryMinInitialDepositRequest) (*v1.QueryMinInitialDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minInitialDeposit := q.k.GetMinInitialDeposit(ctx)

	return &v1.QueryMinInitialDepositResponse{MinInitialDeposit: minInitialDeposit}, nil
}

// Quorums returns the current quorums
func (q queryServer) Quorums(c context.Context, _ *v1.QueryQuorumsRequest) (*v1.QueryQuorumsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &v1.QueryQuorumsResponse{
		Quorum:                      q.k.GetQuorum(ctx).String(),
		ConstitutionAmendmentQuorum: q.k.GetConstitutionAmendmentQuorum(ctx).String(),
		LawQuorum:                   q.k.GetLawQuorum(ctx).String(),
	}, nil
}

// ParticipationEMAs queries the state of the proposal participation exponential moving averages.
func (q queryServer) ParticipationEMAs(c context.Context, _ *v1.QueryParticipationEMAsRequest) (*v1.QueryParticipationEMAsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	participation, err := q.k.ParticipationEMA.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}

	constitutionParticipation, err := q.k.ConstitutionAmendmentParticipationEMA.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}

	lawParticipation, err := q.k.LawParticipationEMA.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}

	return &v1.QueryParticipationEMAsResponse{
		ParticipationEma:                      participation.String(),
		ConstitutionAmendmentParticipationEma: constitutionParticipation.String(),
		LawParticipationEma:                   lawParticipation.String(),
	}, nil
}

// Governor queries governor information based on governor address.
func (q queryServer) Governor(c context.Context, req *v1.QueryGovernorRequest) (*v1.QueryGovernorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.GovernorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "empty governor address")
	}

	governorAddr, err := types.GovernorAddressFromBech32(req.GovernorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	governor, err := q.k.Governors.Get(ctx, governorAddr)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &v1.QueryGovernorResponse{Governor: &governor}, nil
}

// Governors queries all governors.
func (q queryServer) Governors(c context.Context, req *v1.QueryGovernorsRequest) (*v1.QueryGovernorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var governors []*v1.Governor
	governors, pageRes, err := query.CollectionPaginate(ctx, q.k.Governors, req.Pagination, func(_ types.GovernorAddress, governor v1.Governor) (*v1.Governor, error) {
		return &governor, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryGovernorsResponse{Governors: governors, Pagination: pageRes}, nil
}

// GovernanceDelegations queries all delegations of a governor.
func (q queryServer) GovernanceDelegations(c context.Context, req *v1.QueryGovernanceDelegationsRequest) (*v1.QueryGovernanceDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.GovernorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "empty governor address")
	}

	governorAddr, err := types.GovernorAddressFromBech32(req.GovernorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	var delegations []*v1.GovernanceDelegation
	delegations, pageRes, err := query.CollectionPaginate(ctx, q.k.GovernanceDelegationsByGovernor, req.Pagination, func(_ collections.Pair[types.GovernorAddress, sdk.AccAddress], delegation v1.GovernanceDelegation) (*v1.GovernanceDelegation, error) {
		return &delegation, nil
	}, query.WithCollectionPaginationPairPrefix[types.GovernorAddress, sdk.AccAddress](governorAddr))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryGovernanceDelegationsResponse{Delegations: delegations, Pagination: pageRes}, nil
}

// GovernanceDelegation queries a delegation
func (q queryServer) GovernanceDelegation(c context.Context, req *v1.QueryGovernanceDelegationRequest) (*v1.QueryGovernanceDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "empty delegator address")
	}

	delegatorAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	delegation, err := q.k.GovernanceDelegations.Get(ctx, delegatorAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if errors.Is(err, collections.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "governance delegation for %s does not exist", req.DelegatorAddress)
	}

	return &v1.QueryGovernanceDelegationResponse{GovernorAddress: delegation.GovernorAddress}, nil
}

// GovernorValShares queries all validator shares of a governor.
func (q queryServer) GovernorValShares(c context.Context, req *v1.QueryGovernorValSharesRequest) (*v1.QueryGovernorValSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.GovernorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "empty governor address")
	}

	governorAddr, err := types.GovernorAddressFromBech32(req.GovernorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	var valShares []*v1.GovernorValShares
	valShares, pageRes, err := query.CollectionPaginate(ctx, q.k.ValidatorSharesByGovernor, req.Pagination, func(_ collections.Pair[types.GovernorAddress, sdk.ValAddress], valShare v1.GovernorValShares) (*v1.GovernorValShares, error) {
		return &valShare, nil
	}, query.WithCollectionPaginationPairPrefix[types.GovernorAddress, sdk.ValAddress](governorAddr))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryGovernorValSharesResponse{ValShares: valShares, Pagination: pageRes}, nil
}

var _ v1beta1.QueryServer = legacyQueryServer{}

type legacyQueryServer struct{ qs v1.QueryServer }

// NewLegacyQueryServer returns an implementation of the v1beta1 legacy QueryServer interface.
func NewLegacyQueryServer(k *Keeper) v1beta1.QueryServer {
	return &legacyQueryServer{qs: NewQueryServer(k)}
}

func (q legacyQueryServer) Proposal(ctx context.Context, req *v1beta1.QueryProposalRequest) (*v1beta1.QueryProposalResponse, error) {
	resp, err := q.qs.Proposal(ctx, &v1.QueryProposalRequest{
		ProposalId: req.ProposalId,
	})
	if err != nil {
		return nil, err
	}

	proposal, err := v3.ConvertToLegacyProposal(*resp.Proposal)
	if err != nil {
		return nil, err
	}

	return &v1beta1.QueryProposalResponse{Proposal: proposal}, nil
}

func (q legacyQueryServer) Proposals(ctx context.Context, req *v1beta1.QueryProposalsRequest) (*v1beta1.QueryProposalsResponse, error) {
	resp, err := q.qs.Proposals(ctx, &v1.QueryProposalsRequest{
		ProposalStatus: v1.ProposalStatus(req.ProposalStatus),
		Voter:          req.Voter,
		Depositor:      req.Depositor,
		Pagination:     req.Pagination,
	})
	if err != nil {
		return nil, err
	}

	legacyProposals := make([]v1beta1.Proposal, len(resp.Proposals))
	for idx, proposal := range resp.Proposals {
		legacyProposals[idx], err = v3.ConvertToLegacyProposal(*proposal)
		if err != nil {
			return nil, err
		}
	}

	return &v1beta1.QueryProposalsResponse{
		Proposals:  legacyProposals,
		Pagination: resp.Pagination,
	}, nil
}

func (q legacyQueryServer) Vote(ctx context.Context, req *v1beta1.QueryVoteRequest) (*v1beta1.QueryVoteResponse, error) {
	resp, err := q.qs.Vote(ctx, &v1.QueryVoteRequest{
		ProposalId: req.ProposalId,
		Voter:      req.Voter,
	})
	if err != nil {
		return nil, err
	}

	vote, err := v3.ConvertToLegacyVote(*resp.Vote)
	if err != nil {
		return nil, err
	}

	return &v1beta1.QueryVoteResponse{Vote: vote}, nil
}

func (q legacyQueryServer) Votes(ctx context.Context, req *v1beta1.QueryVotesRequest) (*v1beta1.QueryVotesResponse, error) {
	resp, err := q.qs.Votes(ctx, &v1.QueryVotesRequest{
		ProposalId: req.ProposalId,
		Pagination: req.Pagination,
	})
	if err != nil {
		return nil, err
	}

	votes := make([]v1beta1.Vote, len(resp.Votes))
	for i, v := range resp.Votes {
		votes[i], err = v3.ConvertToLegacyVote(*v)
		if err != nil {
			return nil, err
		}
	}

	return &v1beta1.QueryVotesResponse{
		Votes:      votes,
		Pagination: resp.Pagination,
	}, nil
}

//nolint:staticcheck // this is needed for legacy param support
func (q legacyQueryServer) Params(ctx context.Context, req *v1beta1.QueryParamsRequest) (*v1beta1.QueryParamsResponse, error) {
	resp, err := q.qs.Params(ctx, &v1.QueryParamsRequest{
		ParamsType: req.ParamsType,
	})
	if err != nil {
		return nil, err
	}

	response := &v1beta1.QueryParamsResponse{}

	if resp.DepositParams != nil {
		minDeposit := sdk.NewCoins(resp.DepositParams.MinDeposit...)
		response.DepositParams = v1beta1.NewDepositParams(minDeposit, *resp.DepositParams.MaxDepositPeriod)
	}

	if resp.VotingParams != nil {
		response.VotingParams = v1beta1.NewVotingParams(*resp.VotingParams.VotingPeriod)
	}

	if resp.TallyParams != nil {
		quorumRes, err := q.qs.Quorums(ctx, &v1.QueryQuorumsRequest{})
		if err != nil {
			return nil, err
		}
		threshold, err := math.LegacyNewDecFromStr(resp.TallyParams.Threshold)
		if err != nil {
			return nil, err
		}
		quorum := math.LegacyMustNewDecFromStr(quorumRes.Quorum)
		response.TallyParams = v1beta1.NewTallyParams(quorum, threshold, math.LegacyZeroDec())
	}

	return response, nil
}

func (q legacyQueryServer) Deposit(ctx context.Context, req *v1beta1.QueryDepositRequest) (*v1beta1.QueryDepositResponse, error) {
	resp, err := q.qs.Deposit(ctx, &v1.QueryDepositRequest{
		ProposalId: req.ProposalId,
		Depositor:  req.Depositor,
	})
	if err != nil {
		return nil, err
	}

	deposit := v3.ConvertToLegacyDeposit(resp.Deposit)
	return &v1beta1.QueryDepositResponse{Deposit: deposit}, nil
}

func (q legacyQueryServer) Deposits(ctx context.Context, req *v1beta1.QueryDepositsRequest) (*v1beta1.QueryDepositsResponse, error) {
	resp, err := q.qs.Deposits(ctx, &v1.QueryDepositsRequest{
		ProposalId: req.ProposalId,
		Pagination: req.Pagination,
	})
	if err != nil {
		return nil, err
	}
	deposits := make([]v1beta1.Deposit, len(resp.Deposits))
	for idx, deposit := range resp.Deposits {
		deposits[idx] = v3.ConvertToLegacyDeposit(deposit)
	}

	return &v1beta1.QueryDepositsResponse{Deposits: deposits, Pagination: resp.Pagination}, nil
}

func (q legacyQueryServer) TallyResult(ctx context.Context, req *v1beta1.QueryTallyResultRequest) (*v1beta1.QueryTallyResultResponse, error) {
	resp, err := q.qs.TallyResult(ctx, &v1.QueryTallyResultRequest{
		ProposalId: req.ProposalId,
	})
	if err != nil {
		return nil, err
	}

	tally, err := v3.ConvertToLegacyTallyResult(resp.Tally)
	if err != nil {
		return nil, err
	}

	return &v1beta1.QueryTallyResultResponse{Tally: tally}, nil
}
