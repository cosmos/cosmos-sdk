package testutil

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/rest"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

func (s *IntegrationTestSuite) TestGetProposalGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"empty proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s", val.APIAddress, ""),
			true,
		},
		{
			"get non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s", val.APIAddress, "10"),
			true,
		},
		{
			"get proposal with id",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var proposal v1beta2.QueryProposalResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &proposal)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(proposal.Proposal)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetProposalsGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name             string
		url              string
		headers          map[string]string
		wantNumProposals int
		expErr           bool
	}{
		{
			"get proposals with height 1",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals", val.APIAddress),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			0,
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals", val.APIAddress),
			map[string]string{},
			3,
			false,
		},
		{
			"valid request with filter by status",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals?proposal_status=1", val.APIAddress),
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

			var proposals v1beta2.QueryProposalsResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &proposals)

			if tc.expErr {
				s.Require().Empty(proposals.Proposals)
			} else {
				s.Require().NoError(err)
				s.Require().Len(proposals.Proposals, tc.wantNumProposals)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetProposalVoteGRPC() {
	val := s.network.Validators[0]

	voterAddressBech32 := val.Address.String()

	testCases := []struct {
		name           string
		url            string
		expErr         bool
		expVoteOptions v1beta2.WeightedVoteOptions
	}{
		{
			"empty proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/votes/%s", val.APIAddress, "", voterAddressBech32),
			true,
			v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes),
		},
		{
			"get non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/votes/%s", val.APIAddress, "10", voterAddressBech32),
			true,
			v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes),
		},
		{
			"get proposal with wrong voter address",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/votes/%s", val.APIAddress, "1", "wrongVoterAddress"),
			true,
			v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes),
		},
		{
			"get proposal with id",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/votes/%s", val.APIAddress, "1", voterAddressBech32),
			false,
			v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes),
		},
		{
			"get proposal with id for split vote",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/votes/%s", val.APIAddress, "3", voterAddressBech32),
			false,
			v1beta2.WeightedVoteOptions{
				&v1beta2.WeightedVoteOption{Option: v1beta2.OptionYes, Weight: sdk.NewDecWithPrec(60, 2).String()},
				&v1beta2.WeightedVoteOption{Option: v1beta2.OptionNo, Weight: sdk.NewDecWithPrec(30, 2).String()},
				&v1beta2.WeightedVoteOption{Option: v1beta2.OptionAbstain, Weight: sdk.NewDecWithPrec(5, 2).String()},
				&v1beta2.WeightedVoteOption{Option: v1beta2.OptionNoWithVeto, Weight: sdk.NewDecWithPrec(5, 2).String()},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var vote v1beta2.QueryVoteResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &vote)

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

func (s *IntegrationTestSuite) TestGetProposalVotesGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"votes with empty proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/votes", val.APIAddress, ""),
			true,
		},
		{
			"get votes with valid id",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/votes", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var votes v1beta2.QueryVotesResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &votes)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(votes.Votes, 1)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetProposalDepositGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"get deposit with empty proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/deposits/%s", val.APIAddress, "", val.Address.String()),
			true,
		},
		{
			"get deposit of non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/deposits/%s", val.APIAddress, "10", val.Address.String()),
			true,
		},
		{
			"get deposit with wrong depositer address",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/deposits/%s", val.APIAddress, "1", "wrongDepositerAddress"),
			true,
		},
		{
			"get deposit valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/deposits/%s", val.APIAddress, "1", val.Address.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var deposit v1beta2.QueryDepositResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &deposit)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(deposit.Deposit)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetProposalDepositsGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"get deposits with empty proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/deposits", val.APIAddress, ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/deposits", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var deposits v1beta2.QueryDepositsResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &deposits)

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

func (s *IntegrationTestSuite) TestGetTallyGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"get tally with no proposal id",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/tally", val.APIAddress, ""),
			true,
		},
		{
			"get tally with non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/tally", val.APIAddress, "10"),
			true,
		},
		{
			"get tally valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/proposals/%s/tally", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var tally v1beta2.QueryTallyResultResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &tally)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(tally.Tally)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetParamsGRPC() {
	val := s.network.Validators[0]

	dp := v1beta2.DefaultDepositParams()
	vp := v1beta2.DefaultVotingParams()
	tp := v1beta2.DefaultTallyParams()

	testCases := []struct {
		name       string
		url        string
		expErr     bool
		respType   proto.Message
		expectResp proto.Message
	}{
		{
			"request params with empty params type",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/params/%s", val.APIAddress, ""),
			true, nil, nil,
		},
		{
			"get deposit params",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/params/%s", val.APIAddress, v1beta2.ParamDeposit),
			false,
			&v1beta2.QueryParamsResponse{},
			&v1beta2.QueryParamsResponse{DepositParams: &dp},
		},
		{
			"get vote params",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/params/%s", val.APIAddress, v1beta2.ParamVoting),
			false,
			&v1beta2.QueryParamsResponse{},
			&v1beta2.QueryParamsResponse{VotingParams: &vp},
		},
		{
			"get tally params",
			fmt.Sprintf("%s/cosmos/gov/v1beta2/params/%s", val.APIAddress, v1beta2.ParamTallying),
			false,
			&v1beta2.QueryParamsResponse{},
			&v1beta2.QueryParamsResponse{TallyParams: &tp},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectResp.String(), tc.respType.String())
			}
		})
	}
}
