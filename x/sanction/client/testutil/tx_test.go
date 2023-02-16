package testutil

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	client "github.com/cosmos/cosmos-sdk/x/sanction/client/cli"
)

// assertGovPropMsg gets a gov prop and makes sure it has one specific message.
func (s *IntegrationTestSuite) assertGovPropMsg(propID string, msg sdk.Msg) bool {
	s.T().Helper()
	if msg == nil {
		return true
	}

	if !s.Assert().NotEmpty(propID, "proposal id") {
		return false
	}
	expPropMsgAny, err := codectypes.NewAnyWithValue(msg)
	if !s.Assert().NoError(err, "NewAnyWithValue on %T", msg) {
		return false
	}

	getPropCmd := govcli.GetCmdQueryProposal()
	propOutBW, err := cli.ExecTestCLICmd(s.clientCtx, getPropCmd, []string{propID, "--output", "json"})
	propOut := propOutBW.String()
	s.T().Logf("Query proposal %s output:\n%s", propID, propOut)
	if !s.Assert().NoError(err, "GetCmdQueryProposal error") {
		return false
	}

	var prop govv1.Proposal
	err = s.clientCtx.Codec.UnmarshalJSON([]byte(propOut), &prop)
	if !s.Assert().NoError(err, "UnmarshalJSON on proposal response") {
		return false
	}
	if !s.Assert().Len(prop.Messages, 1, "number of messages in proposal") {
		return false
	}
	if !s.Assert().Equal(expPropMsgAny, prop.Messages[0], "the message in the proposal") {
		return false
	}

	return true
}

// findProposalID looks through the provided response to find a governance proposal id.
// If one is found, it's returned (as a string). Otherwise, an empty string is returned.
func (s *IntegrationTestSuite) findProposalID(resp *sdk.TxResponse) string {
	for _, event := range resp.Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "proposal_id" {
					return string(attr.Value)
				}
			}
		}
	}
	return ""
}

func (s *IntegrationTestSuite) TestTxSanctionCmd() {
	authority := s.getAuthority()
	addr1 := sdk.AccAddress("1_address_test_test_").String()
	addr2 := sdk.AccAddress("2_address_test_test_").String()

	tests := []struct {
		name       string
		args       []string
		expErr     []string
		expPropMsg *sanction.MsgSanction
	}{
		{
			name:   "no addresses given",
			args:   []string{},
			expErr: []string{"requires at least 1 arg(s), only received 0"},
		},
		{
			name: "one address good",
			args: []string{addr1},
			expPropMsg: &sanction.MsgSanction{
				Addresses: []string{addr1},
				Authority: authority,
			},
		},
		{
			name:   "one address bad",
			args:   []string{"thisis1addrthatisbad"},
			expErr: []string{"addresses[0]", `"thisis1addrthatisbad"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses first bad",
			args:   []string{"another1badaddr", addr2},
			expErr: []string{"addresses[0]", `"another1badaddr"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses second bad",
			args:   []string{addr1, "athird1badaddress"},
			expErr: []string{"addresses[1]", `"athird1badaddress"`, "decoding bech32 failed"},
		},
		{
			name: "two addresses good",
			args: []string{addr1, addr2},
			expPropMsg: &sanction.MsgSanction{
				Addresses: []string{addr1, addr2},
				Authority: authority,
			},
		},
		{
			name:   "bad authority",
			args:   []string{addr1, "--" + flags.FlagAuthority, "bad1auth34sd2"},
			expErr: []string{"authority", `"bad1auth34sd2"`, "decoding bech32 failed"},
		},
		{
			name:   "bad deposit",
			args:   []string{addr1, "--" + govcli.FlagDeposit, "notcoins"},
			expErr: []string{"invalid deposit", "notcoins"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.TxSanctionCmd()
			cmdFuncName := "TxSanctionCmd"
			args := s.appendCommonArgsTo(tc.args...)

			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			out := outBW.String()
			s.T().Logf("Output:\n%s", out)
			s.assertErrorContents(err, tc.expErr, "%s error", cmdFuncName)
			for _, expErr := range tc.expErr {
				s.Assert().Contains(out, expErr, "%s output with error", cmdFuncName)
			}

			var propID string
			if len(tc.expErr) == 0 {
				var txResp sdk.TxResponse
				err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), &txResp)
				if s.Assert().NoError(err, "UnmarshalJSON on %s", cmdFuncName) {
					s.Assert().Equal(0, int(txResp.Code), "%s response code", cmdFuncName)
				}
				propID = s.findProposalID(&txResp)
			}

			if tc.expPropMsg != nil {
				s.assertGovPropMsg(propID, tc.expPropMsg)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUnsanctionCmd() {
	authority := s.getAuthority()
	addr1 := sdk.AccAddress("1_address_untest____").String()
	addr2 := sdk.AccAddress("2_address_untest____").String()

	tests := []struct {
		name       string
		args       []string
		expErr     []string
		expPropMsg *sanction.MsgUnsanction
	}{
		{
			name:   "no addresses given",
			args:   []string{},
			expErr: []string{"requires at least 1 arg(s), only received 0"},
		},
		{
			name: "one address good",
			args: []string{addr1},
			expPropMsg: &sanction.MsgUnsanction{
				Addresses: []string{addr1},
				Authority: authority,
			},
		},
		{
			name:   "one address bad",
			args:   []string{"thisis1addrthatisbad"},
			expErr: []string{"addresses[0]", `"thisis1addrthatisbad"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses first bad",
			args:   []string{"another1badaddr", addr2},
			expErr: []string{"addresses[0]", `"another1badaddr"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses second bad",
			args:   []string{addr1, "athird1badaddress"},
			expErr: []string{"addresses[1]", `"athird1badaddress"`, "decoding bech32 failed"},
		},
		{
			name: "two addresses good",
			args: []string{addr1, addr2},
			expPropMsg: &sanction.MsgUnsanction{
				Addresses: []string{addr1, addr2},
				Authority: authority,
			},
		},
		{
			name:   "bad authority",
			args:   []string{addr1, "--" + flags.FlagAuthority, "bad1auth34sd2"},
			expErr: []string{"authority", `"bad1auth34sd2"`, "decoding bech32 failed"},
		},
		{
			name:   "bad deposit",
			args:   []string{addr1, "--" + govcli.FlagDeposit, "notcoins"},
			expErr: []string{"invalid deposit", "notcoins"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.TxUnsanctionCmd()
			cmdFuncName := "TxUnsanctionCmd"
			args := s.appendCommonArgsTo(tc.args...)

			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			out := outBW.String()
			s.T().Logf("Output:\n%s", out)
			s.assertErrorContents(err, tc.expErr, "%s error", cmdFuncName)
			for _, expErr := range tc.expErr {
				s.Assert().Contains(out, expErr, "%s output with error", cmdFuncName)
			}

			var propID string
			if len(tc.expErr) == 0 {
				var txResp sdk.TxResponse
				err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), &txResp)
				if s.Assert().NoError(err, "UnmarshalJSON on %s", cmdFuncName) {
					s.Assert().Equal(0, int(txResp.Code), "%s response code", cmdFuncName)
				}
				propID = s.findProposalID(&txResp)
			}

			if tc.expPropMsg != nil {
				s.assertGovPropMsg(propID, tc.expPropMsg)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateParamsCmd() {
	authority := s.getAuthority()

	tests := []struct {
		name       string
		args       []string
		expErr     []string
		expPropMsg *sanction.MsgUpdateParams
	}{
		{
			name:   "no args",
			args:   []string{},
			expErr: []string{"accepts 2 arg(s), received 0"},
		},
		{
			name:   "one arg",
			args:   []string{"arg1"},
			expErr: []string{"accepts 2 arg(s), received 1"},
		},
		{
			name:   "three args",
			args:   []string{"arg1", "arg2", "arg3"},
			expErr: []string{"accepts 2 arg(s), received 3"},
		},
		{
			name: "coins coins",
			args: []string{"1acoin", "2bcoin"},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("acoin", 1)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bcoin", 2)),
				},
				Authority: authority,
			},
		},
		{
			name: "empty coins",
			args: []string{"", "3ccoin"},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("ccoin", 3)),
				},
				Authority: authority,
			},
		},
		{
			name: "coins empty",
			args: []string{"4dcoin", ""},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("dcoin", 4)),
					ImmediateUnsanctionMinDeposit: nil,
				},
				Authority: authority,
			},
		},
		{
			name: "empty empty",
			args: []string{"", ""},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: nil,
				},
				Authority: authority,
			},
		},
		{
			name:   "bad good",
			args:   []string{"firscoinsbad", "5ecoin"},
			expErr: []string{"invalid immediate_sanction_min_deposit", `"firscoinsbad"`},
		},
		{
			name:   "good bad",
			args:   []string{"6fcoin", "secondcoinsbad"},
			expErr: []string{"invalid immediate_unsanction_min_deposit", `"secondcoinsbad"`},
		},
		{
			name:   "bad authority",
			args:   []string{"", "", "--" + flags.FlagAuthority, "bad1auth34sd2"},
			expErr: []string{"authority", `"bad1auth34sd2"`, "decoding bech32 failed"},
		},
		{
			name:   "bad deposit",
			args:   []string{"", "", "--" + govcli.FlagDeposit, "notcoins"},
			expErr: []string{"invalid deposit", "notcoins"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.TxUpdateParamsCmd()
			cmdFuncName := "TxUpdateParamsCmd"
			args := s.appendCommonArgsTo(tc.args...)

			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			out := outBW.String()
			s.T().Logf("Output:\n%s", out)
			s.assertErrorContents(err, tc.expErr, "%s error", cmdFuncName)
			for _, expErr := range tc.expErr {
				s.Assert().Contains(out, expErr, "%s output with error", cmdFuncName)
			}

			var propID string
			if len(tc.expErr) == 0 {
				var txResp sdk.TxResponse
				err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), &txResp)
				if s.Assert().NoError(err, "UnmarshalJSON on %s", cmdFuncName) {
					s.Assert().Equal(0, int(txResp.Code), "%s response code", cmdFuncName)
				}
				propID = s.findProposalID(&txResp)
			}

			if tc.expPropMsg != nil {
				s.assertGovPropMsg(propID, tc.expPropMsg)
			}
		})
	}
}
