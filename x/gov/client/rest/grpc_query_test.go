// +build norace

package rest_test

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = network.DefaultConfig()
	s.cfg.NumValidators = 1

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create a proposal with deposit
	_, err = govtestutil.MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", types.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// vote for proposal
	_, err = govtestutil.MsgVote(val.ClientCtx, val.Address.String(), "1", "yes")
	s.Require().NoError(err)

	// create a proposal without deposit
	_, err = govtestutil.MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 2", "Where is the title!?", types.ProposalTypeText)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestGetProposalGRPC() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		url    string
		expErr bool
	}{
		{
			"empty proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s", val.APIAddress, ""),
			true,
		},
		{
			"get non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s", val.APIAddress, "10"),
			true,
		},
		{
			"get proposal with id",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var proposal types.QueryProposalResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &proposal)

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
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals", val.APIAddress),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			0,
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals", val.APIAddress),
			map[string]string{},
			2,
			false,
		},
		{
			"valid request with filter by status",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals?proposal_status=1", val.APIAddress),
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

			var proposals types.QueryProposalsResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &proposals)

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
		name   string
		url    string
		expErr bool
	}{
		{
			"empty proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/votes/%s", val.APIAddress, "", voterAddressBech32),
			true,
		},
		{
			"get non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/votes/%s", val.APIAddress, "10", voterAddressBech32),
			true,
		},
		{
			"get proposal with wrong voter address",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/votes/%s", val.APIAddress, "1", "wrongVoterAddress"),
			true,
		},
		{
			"get proposal with id",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/votes/%s", val.APIAddress, "1", voterAddressBech32),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var vote types.QueryVoteResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &vote)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(vote.Vote)
				s.Require().Equal(types.OptionYes, vote.Vote.Option)
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
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/votes", val.APIAddress, ""),
			true,
		},
		{
			"get votes with valid id",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/votes", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var votes types.QueryVotesResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &votes)

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
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/deposits/%s", val.APIAddress, "", val.Address.String()),
			true,
		},
		{
			"get deposit of non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/deposits/%s", val.APIAddress, "10", val.Address.String()),
			true,
		},
		{
			"get deposit with wrong depositer address",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/deposits/%s", val.APIAddress, "1", "wrongDepositerAddress"),
			true,
		},
		{
			"get deposit valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/deposits/%s", val.APIAddress, "1", val.Address.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var deposit types.QueryDepositResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &deposit)

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
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/deposits", val.APIAddress, ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/deposits", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var deposits types.QueryDepositsResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &deposits)

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
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/tally", val.APIAddress, ""),
			true,
		},
		{
			"get tally with non existing proposal",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/tally", val.APIAddress, "10"),
			true,
		},
		{
			"get tally valid request",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%s/tally", val.APIAddress, "1"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var tally types.QueryTallyResultResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &tally)

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

	testCases := []struct {
		name       string
		url        string
		expErr     bool
		respType   proto.Message
		expectResp proto.Message
	}{
		{
			"request params with empty params type",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/params/%s", val.APIAddress, ""),
			true, nil, nil,
		},
		{
			"get deposit params",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/params/%s", val.APIAddress, types.ParamDeposit),
			false,
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				DepositParams: types.DefaultDepositParams(),
			},
		},
		{
			"get vote params",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/params/%s", val.APIAddress, types.ParamVoting),
			false,
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				VotingParams: types.DefaultVotingParams(),
			},
		},
		{
			"get tally params",
			fmt.Sprintf("%s/cosmos/gov/v1beta1/params/%s", val.APIAddress, types.ParamTallying),
			false,
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				TallyParams: types.DefaultTallyParams(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectResp.String(), tc.respType.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
