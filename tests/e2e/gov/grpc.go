package gov

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/math"
	v1 "cosmossdk.io/x/gov/types/v1"

	"github.com/cosmos/cosmos-sdk/testutil"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

func (s *E2ETestSuite) TestGetProposalGRPC() {
	val := s.network.GetValidators()[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"empty proposal",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", val.GetAPIAddress(), ""),
			true,
		},
		{
			"get non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", val.GetAPIAddress(), "10"),
			true,
		},
		{
			"get proposal with id",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", val.GetAPIAddress(), "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)

			var proposal v1.QueryProposalResponse
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, &proposal)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(proposal.Proposal)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetProposalsGRPC() {
	val := s.network.GetValidators()[0]

	testCases := []struct {
		name             string
		url              string
		headers          map[string]string
		wantNumProposals int
		expErr           bool
	}{
		{
			"get proposals with height 1",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals", val.GetAPIAddress()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			0,
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals", val.GetAPIAddress()),
			map[string]string{},
			4,
			false,
		},
		{
			"valid request with filter by status",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals?proposal_status=1", val.GetAPIAddress()),
			map[string]string{},
			1,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var proposals v1.QueryProposalsResponse
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, &proposals)

			if tc.expErr {
				s.Require().Empty(proposals.Proposals)
			} else {
				s.Require().NoError(err)
				s.Require().Len(proposals.Proposals, tc.wantNumProposals)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetProposalVoteGRPC() {
	val := s.network.GetValidators()[0]

	voterAddressBech32 := val.GetAddress().String()

	testCases := []struct {
		name           string
		url            string
		expErr         bool
		expVoteOptions v1.WeightedVoteOptions
	}{
		{
			"empty proposal",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes/%s", val.GetAPIAddress(), "", voterAddressBech32),
			true,
			v1.NewNonSplitVoteOption(v1.OptionYes),
		},
		{
			"get non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes/%s", val.GetAPIAddress(), "10", voterAddressBech32),
			true,
			v1.NewNonSplitVoteOption(v1.OptionYes),
		},
		{
			"get proposal with wrong voter address",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes/%s", val.GetAPIAddress(), "1", "wrongVoterAddress"),
			true,
			v1.NewNonSplitVoteOption(v1.OptionYes),
		},
		{
			"get proposal with id",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes/%s", val.GetAPIAddress(), "1", voterAddressBech32),
			false,
			v1.NewNonSplitVoteOption(v1.OptionYes),
		},
		{
			"get proposal with id for split vote",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes/%s", val.GetAPIAddress(), "3", voterAddressBech32),
			false,
			v1.WeightedVoteOptions{
				&v1.WeightedVoteOption{Option: v1.OptionYes, Weight: math.LegacyNewDecWithPrec(60, 2).String()},
				&v1.WeightedVoteOption{Option: v1.OptionNo, Weight: math.LegacyNewDecWithPrec(30, 2).String()},
				&v1.WeightedVoteOption{Option: v1.OptionAbstain, Weight: math.LegacyNewDecWithPrec(5, 2).String()},
				&v1.WeightedVoteOption{Option: v1.OptionNoWithVeto, Weight: math.LegacyNewDecWithPrec(5, 2).String()},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)

			var vote v1.QueryVoteResponse
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, &vote)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(vote.Vote)
				s.Require().Equal(len(vote.Vote.Options), len(tc.expVoteOptions))
				for i, option := range tc.expVoteOptions {
					s.Require().Equal(option.Option, vote.Vote.Options[i].Option)
					s.Require().Equal(option.Weight, vote.Vote.Options[i].Weight)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestGetProposalVotesGRPC() {
	val := s.network.GetValidators()[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"votes with empty proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes", val.GetAPIAddress(), ""),
			true,
		},
		{
			"get votes with valid id",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes", val.GetAPIAddress(), "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)

			var votes v1.QueryVotesResponse
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, &votes)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(votes.Votes, 1)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetProposalDepositGRPC() {
	val := s.network.GetValidators()[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"get deposit with empty proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits/%s", val.GetAPIAddress(), "", val.GetAddress().String()),
			true,
		},
		{
			"get deposit of non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits/%s", val.GetAPIAddress(), "10", val.GetAddress().String()),
			true,
		},
		{
			"get deposit with wrong depositer address",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits/%s", val.GetAPIAddress(), "1", "wrongDepositerAddress"),
			true,
		},
		{
			"get deposit valid request",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits/%s", val.GetAPIAddress(), "1", val.GetAddress().String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)

			var deposit v1.QueryDepositResponse
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, &deposit)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(deposit.Deposit)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetProposalDepositsGRPC() {
	val := s.network.GetValidators()[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"get deposits with empty proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits", val.GetAPIAddress(), ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits", val.GetAPIAddress(), "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)

			var deposits v1.QueryDepositsResponse
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, &deposits)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(deposits.Deposits, 1)
				s.Require().NotEmpty(deposits.Deposits[0])
			}
		})
	}
}

func (s *E2ETestSuite) TestGetTallyGRPC() {
	val := s.network.GetValidators()[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"get tally with no proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/tally", val.GetAPIAddress(), ""),
			true,
		},
		{
			"get tally with non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/tally", val.GetAPIAddress(), "10"),
			true,
		},
		{
			"get tally valid request",
			fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/tally", val.GetAPIAddress(), "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)

			var tally v1.QueryTallyResultResponse
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, &tally)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(tally.Tally)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetParamsGRPC() {
	val := s.network.GetValidators()[0]

	params := v1.DefaultParams()
	dp := v1.NewDepositParams(params.MinDeposit, params.MaxDepositPeriod)          //nolint:staticcheck // we use deprecated gov commands here, but we don't want to remove them
	vp := v1.NewVotingParams(params.VotingPeriod)                                  //nolint:staticcheck // we use deprecated gov commands here, but we don't want to remove them
	tp := v1.NewTallyParams(params.Quorum, params.Threshold, params.VetoThreshold) //nolint:staticcheck // we use deprecated gov commands here, but we don't want to remove them

	testCases := []struct {
		name       string
		url        string
		expErr     bool
		respType   proto.Message
		expectResp proto.Message
	}{
		{
			"request params with empty params type",
			fmt.Sprintf("%s/cosmos/gov/v1/params/%s", val.GetAPIAddress(), ""),
			true, nil, nil,
		},
		{
			"get deposit params",
			fmt.Sprintf("%s/cosmos/gov/v1/params/%s", val.GetAPIAddress(), v1.ParamDeposit),
			false,
			&v1.QueryParamsResponse{},
			&v1.QueryParamsResponse{DepositParams: &dp, Params: &params},
		},
		{
			"get vote params",
			fmt.Sprintf("%s/cosmos/gov/v1/params/%s", val.GetAPIAddress(), v1.ParamVoting),
			false,
			&v1.QueryParamsResponse{},
			&v1.QueryParamsResponse{VotingParams: &vp, Params: &params},
		},
		{
			"get tally params",
			fmt.Sprintf("%s/cosmos/gov/v1/params/%s", val.GetAPIAddress(), v1.ParamTallying),
			false,
			&v1.QueryParamsResponse{},
			&v1.QueryParamsResponse{TallyParams: &tp, Params: &params},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)
			err = val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectResp.String(), tc.respType.String())
			}
		})
	}
}
