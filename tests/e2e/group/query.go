package group

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli"
)

func (s *E2ETestSuite) TestTallyResult() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	member := s.voter

	// create a proposal
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
		append(
			[]string{
				s.createCLIProposal(
					s.groupPolicies[0].Address, val.Address.String(),
					s.groupPolicies[0].Address, val.Address.String(),
					"", "title", "summary"),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())

	var txResp sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	txResp, err = clitestutil.GetTxResponse(s.network, clientCtx, txResp.TxHash)
	s.Require().NoError(err)
	s.Require().Equal(txResp.Code, uint32(0), out.String())

	proposalID := s.getProposalIDFromTxResponse(txResp)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expTallyResult group.TallyResult
		expectErrMsg   string
	}{
		{
			"not found",
			[]string{
				"12345",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			group.TallyResult{},
			"not found",
		},
		{
			"invalid proposal id",
			[]string{
				"",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			group.TallyResult{},
			"strconv.ParseUint: parsing \"\": invalid syntax",
		},
		{
			"valid proposal id with no votes",
			[]string{
				proposalID,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			group.DefaultTallyResult(),
			"",
		},
		{
			"valid proposal id",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			group.TallyResult{
				YesCount:        member.Weight,
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryTallyResultCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				var tallyResultRes group.QueryTallyResultResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &tallyResultRes))
				s.Require().NotNil(tallyResultRes)
				s.Require().Equal(tc.expTallyResult, tallyResultRes.Tally)
			}
		})
	}
}
