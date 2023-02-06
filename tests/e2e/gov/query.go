package gov

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (s *E2ETestSuite) TestCmdParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			`{"voting_params":{"voting_period":"172800s"},"deposit_params":{"min_deposit":[{"denom":"stake","amount":"10000000"}],"max_deposit_period":"172800s"},"tally_params":{"quorum":"0.334000000000000000","threshold":"0.500000000000000000","veto_threshold":"0.334000000000000000"},"params":{"min_deposit":[{"denom":"stake","amount":"10000000"}],"max_deposit_period":"172800s","voting_period":"172800s","quorum":"0.334000000000000000","threshold":"0.500000000000000000","veto_threshold":"0.334000000000000000","min_initial_deposit_ratio":"0.000000000000000000","proposal_cancel_ratio":"0.500000000000000000","proposal_cancel_dest":"","expedited_voting_period":"86400s","expedited_threshold":"0.667000000000000000","expedited_min_deposit":[{"denom":"stake","amount":"50000000"}]}}`,
		},
		{
			"text output",
			[]string{},
			`
deposit_params:
  max_deposit_period: 172800s
  min_deposit:
  - amount: "10000000"
    denom: stake
params:
  expedited_min_deposit:
  - amount: "50000000"
    denom: stake
  expedited_threshold: "0.667000000000000000"
  expedited_voting_period: 86400s
  max_deposit_period: 172800s
  min_deposit:
  - amount: "10000000"
    denom: stake
  min_initial_deposit_ratio: "0.000000000000000000"
  proposal_cancel_dest: ""
  proposal_cancel_ratio: "0.500000000000000000"
  quorum: "0.334000000000000000"
  threshold: "0.500000000000000000"
  veto_threshold: "0.334000000000000000"
  voting_period: 172800s
tally_params:
  quorum: "0.334000000000000000"
  threshold: "0.500000000000000000"
  veto_threshold: "0.334000000000000000"
voting_params:
  voting_period: 172800s
	`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
		})
	}
}

func (s *E2ETestSuite) TestCmdParam() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"voting params",
			[]string{
				"voting",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			`{"voting_period":"172800000000000"}`,
		},
		{
			"tally params",
			[]string{
				"tallying",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			`{"quorum":"0.334000000000000000","threshold":"0.500000000000000000","veto_threshold":"0.334000000000000000"}`,
		},
		{
			"deposit params",
			[]string{
				"deposit",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			`{"min_deposit":[{"denom":"stake","amount":"10000000"}],"max_deposit_period":"172800000000000"}`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParam()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
		})
	}
}

func (s *E2ETestSuite) TestCmdProposer() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"without proposal id",
			[]string{
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			``,
		},
		{
			"json output",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			fmt.Sprintf("{\"proposal_id\":\"%s\",\"proposer\":\"%s\"}", "1", val.Address.String()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryProposer()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
			}
		})
	}
}

func (s *E2ETestSuite) TestCmdTally() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput v1.TallyResult
	}{
		{
			"without proposal id",
			[]string{
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			v1.TallyResult{},
		},
		{
			"json output",
			[]string{
				"2",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			v1.NewTallyResult(sdk.NewInt(0), sdk.NewInt(0), sdk.NewInt(0), sdk.NewInt(0)),
		},
		{
			"json output",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			v1.NewTallyResult(s.cfg.BondedTokens, sdk.NewInt(0), sdk.NewInt(0), sdk.NewInt(0)),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryTally()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var tally v1.TallyResult
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &tally), out.String())
				s.Require().Equal(tally, tc.expectedOutput)
			}
		})
	}
}

func (s *E2ETestSuite) TestCmdGetProposal() {
	val := s.network.Validators[0]

	title := "Text Proposal 1"

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"get non existing proposal",
			[]string{
				"10",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"get proposal with json response",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryProposal()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var proposal v1.Proposal
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposal), out.String())
				s.Require().Equal(title, proposal.Messages[0].GetCachedValue().(*v1.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content).GetTitle())
			}
		})
	}
}

func (s *E2ETestSuite) TestCmdGetProposals() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"get proposals as json response",
			[]string{
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
		{
			"get proposals with invalid status",
			[]string{
				"--status=unknown",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryProposals()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var proposals v1.QueryProposalsResponse

				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposals), out.String())
				s.Require().Greater(len(proposals.Proposals), 0)
			}
		})
	}
}

func (s *E2ETestSuite) TestCmdQueryDeposits() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"get deposits of non existing proposal",
			[]string{
				"10",
			},
			true,
		},
		{
			"get deposits(valid req)",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDeposits()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				var deposits v1.QueryDepositsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &deposits), out.String())
				s.Require().Len(deposits.Deposits, 1)
			}
		})
	}
}

func (s *E2ETestSuite) TestCmdQueryDeposit() {
	val := s.network.Validators[0]
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"get deposit with no depositer",
			[]string{
				"1",
			},
			true,
		},
		{
			"get deposit with wrong deposit address",
			[]string{
				"1",
				"wrong address",
			},
			true,
		},
		{
			"get deposit (valid req)",
			[]string{
				"1",
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDeposit()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				var deposit v1.Deposit
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &deposit), out.String())
				s.Require().Equal(depositAmount.String(), sdk.Coins(deposit.Amount).String())
			}
		})
	}
}

func (s *E2ETestSuite) TestCmdQueryVotes() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"get votes with no proposal id",
			[]string{},
			true,
		},
		{
			"get votes of non existed proposal",
			[]string{
				"10",
			},
			true,
		},
		{
			"vote for invalid proposal",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryVotes()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				var votes v1.QueryVotesResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &votes), out.String())
				s.Require().Len(votes.Votes, 1)
			}
		})
	}
}

func (s *E2ETestSuite) TestCmdQueryVote() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expVoteOptions v1.WeightedVoteOptions
	}{
		{
			"get vote of non existing proposal",
			[]string{
				"10",
				val.Address.String(),
			},
			true,
			v1.NewNonSplitVoteOption(v1.OptionYes),
		},
		{
			"get vote by wrong voter",
			[]string{
				"1",
				"wrong address",
			},
			true,
			v1.NewNonSplitVoteOption(v1.OptionYes),
		},
		{
			"vote for valid proposal",
			[]string{
				"1",
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			v1.NewNonSplitVoteOption(v1.OptionYes),
		},
		{
			"split vote for valid proposal",
			[]string{
				"3",
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			v1.WeightedVoteOptions{
				&v1.WeightedVoteOption{Option: v1.OptionYes, Weight: sdk.NewDecWithPrec(60, 2).String()},
				&v1.WeightedVoteOption{Option: v1.OptionNo, Weight: sdk.NewDecWithPrec(30, 2).String()},
				&v1.WeightedVoteOption{Option: v1.OptionAbstain, Weight: sdk.NewDecWithPrec(5, 2).String()},
				&v1.WeightedVoteOption{Option: v1.OptionNoWithVeto, Weight: sdk.NewDecWithPrec(5, 2).String()},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryVote()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				var vote v1.Vote
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &vote), out.String())
				s.Require().Equal(len(vote.Options), len(tc.expVoteOptions))
				for i, option := range tc.expVoteOptions {
					s.Require().Equal(option.Option, vote.Options[i].Option)
					s.Require().Equal(option.Weight, vote.Options[i].Weight)
				}
			}
		})
	}
}
