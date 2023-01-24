package testutil

import (
	"fmt"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	client "github.com/cosmos/cosmos-sdk/x/sanction/client/cli"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
)

// These tests are initiated by TestIntegrationTestSuite in cli_test.go

func (s *IntegrationTestSuite) TestIsSanctionedCmd() {
	otherAddr := sdk.AccAddress("1_other_test_address")

	tests := []struct {
		name   string
		args   []string
		exp    *sanction.QueryIsSanctionedResponse
		expErr []string
	}{
		{
			name: "not sanctioned",
			args: []string{otherAddr.String()},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: false},
		},
		{
			name: "sanctioned address 1",
			args: []string{s.sanctionGenesis.SanctionedAddresses[0]},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: true},
		},
		{
			name: "sanctioned address 2",
			args: []string{s.sanctionGenesis.SanctionedAddresses[1]},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: true},
		},
		{
			name: "temp sanctioned address",
			args: []string{s.sanctionGenesis.TemporaryEntries[0].Address},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: true},
		},
		{
			name: "temp unsanctioned address",
			args: []string{s.sanctionGenesis.TemporaryEntries[1].Address},
			exp:  &sanction.QueryIsSanctionedResponse{IsSanctioned: false},
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: []string{"accepts 1 arg(s), received 0"},
		},
		{
			name:   "two args",
			args:   []string{"arg1", "arg2"},
			expErr: []string{"accepts 1 arg(s), received 2"},
		},
		{
			name:   "not an address",
			args:   []string{"notanaddress"},
			expErr: []string{"decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.QueryIsSanctionedCmd()
			args := append(tc.args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			outBz := outBW.Bytes()
			s.T().Logf("Output:\n%s", string(outBz))
			s.assertErrorContents(err, tc.expErr, "QueryIsSanctionedCmd error")
			for _, expErr := range tc.expErr {
				s.Assert().Contains(string(outBz), expErr, "QueryIsSanctionedCmd output with error")
			}
			if tc.exp != nil {
				act := &sanction.QueryIsSanctionedResponse{}
				testFunc := func() {
					err = s.clientCtx.Codec.UnmarshalJSON(outBz, act)
				}
				if s.Assert().NotPanics(testFunc, "UnmarshalJSON on output") {
					if s.Assert().NoError(err, "UnmarshalJSON on output") {
						s.Assert().Equal(tc.exp, act, "QueryIsSanctionedCmd response")
					}
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQuerySanctionedAddressesCmd() {
	addr2Key := address.MustLengthPrefix(sdk.MustAccAddressFromBech32(s.sanctionGenesis.SanctionedAddresses[1]))

	tests := []struct {
		name   string
		args   []string
		exp    *sanction.QuerySanctionedAddressesResponse
		expErr []string
	}{
		{
			name:   "arg provided",
			args:   []string{"arg1"},
			expErr: []string{"accepts 0 arg(s), received 1"},
		},
		{
			name: "no args",
			args: []string{},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: s.sanctionGenesis.SanctionedAddresses,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "limit 1",
			args: []string{"--limit", "1"},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{s.sanctionGenesis.SanctionedAddresses[0]},
				Pagination: &query.PageResponse{
					NextKey: addr2Key,
					Total:   0,
				},
			},
		},
		{
			name: "limit 1 offset 1",
			args: []string{"--limit", "1", "--offset", "1"},
			exp: &sanction.QuerySanctionedAddressesResponse{
				Addresses: []string{s.sanctionGenesis.SanctionedAddresses[1]},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.QuerySanctionedAddressesCmd()
			args := append(tc.args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			outBz := outBW.Bytes()
			s.T().Logf("Output:\n%s", string(outBz))
			s.assertErrorContents(err, tc.expErr, "QuerySanctionedAddressesCmd error")
			for _, expErr := range tc.expErr {
				s.Assert().Contains(string(outBz), expErr, "QuerySanctionedAddressesCmd output with error")
			}
			if tc.exp != nil {
				act := &sanction.QuerySanctionedAddressesResponse{}
				testFunc := func() {
					err = s.clientCtx.Codec.UnmarshalJSON(outBz, act)
				}
				if s.Assert().NotPanics(testFunc, "UnmarshalJSON on output") {
					if s.Assert().NoError(err, "UnmarshalJSON on output") {
						s.Assert().Equal(tc.exp.Addresses, act.Addresses, "Addresses")
						s.Assert().Equal(tc.exp.Pagination, act.Pagination, "Pagination")
					}
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryTemporaryEntriesCmd() {
	entry2Key := keeper.CreateTemporaryKey(sdk.MustAccAddressFromBech32(s.sanctionGenesis.TemporaryEntries[1].Address), s.sanctionGenesis.TemporaryEntries[1].ProposalId)[1:]

	tests := []struct {
		name   string
		args   []string
		exp    *sanction.QueryTemporaryEntriesResponse
		expErr []string
	}{
		{
			name:   "two args provided",
			args:   []string{"arg1", "arg2"},
			expErr: []string{"accepts at most 1 arg(s), received 2"},
		},
		{
			name:   "not an address",
			args:   []string{"notanaddress"},
			expErr: []string{"decoding bech32 failed"},
		},
		{
			name: "no args",
			args: []string{},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: s.sanctionGenesis.TemporaryEntries,
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "limit 1",
			args: []string{"--limit", "1"},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					s.sanctionGenesis.TemporaryEntries[0],
				},
				Pagination: &query.PageResponse{
					NextKey: entry2Key,
					Total:   0,
				},
			},
		},
		{
			name: "limit 1 offset 1",
			args: []string{"--limit", "1", "--offset", "1"},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					s.sanctionGenesis.TemporaryEntries[1],
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "addr provided with an entry",
			args: []string{s.sanctionGenesis.TemporaryEntries[0].Address},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{
					s.sanctionGenesis.TemporaryEntries[0],
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
		{
			name: "addr provided without entries",
			args: []string{s.sanctionGenesis.SanctionedAddresses[0]},
			exp: &sanction.QueryTemporaryEntriesResponse{
				Entries: []*sanction.TemporaryEntry{},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   0,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.QueryTemporaryEntriesCmd()
			args := append(tc.args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			outBz := outBW.Bytes()
			s.T().Logf("Output:\n%s", string(outBz))
			s.assertErrorContents(err, tc.expErr, "QueryTemporaryEntriesCmd error")
			for _, expErr := range tc.expErr {
				s.Assert().Contains(string(outBz), expErr, "QueryTemporaryEntriesCmd output with error")
			}
			if tc.exp != nil {
				act := &sanction.QueryTemporaryEntriesResponse{}
				testFunc := func() {
					err = s.clientCtx.Codec.UnmarshalJSON(outBz, act)
				}
				if s.Assert().NotPanics(testFunc, "UnmarshalJSON on output") {
					if s.Assert().NoError(err, "UnmarshalJSON on output") {
						s.Assert().Equal(tc.exp.Entries, act.Entries, "Entries")
						s.Assert().Equal(tc.exp.Pagination, act.Pagination, "Pagination")
					}
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryParamsCmd() {
	tests := []struct {
		name   string
		args   []string
		exp    *sanction.QueryParamsResponse
		expErr []string
	}{
		{
			name: "no args",
			args: []string{},
			exp:  &sanction.QueryParamsResponse{Params: s.sanctionGenesis.Params},
		},
		{
			name:   "one arg",
			args:   []string{"arg1"},
			expErr: []string{"accepts 0 arg(s), received 1"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.QueryParamsCmd()
			args := append(tc.args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			outBz := outBW.Bytes()
			s.T().Logf("Output:\n%s", string(outBz))
			s.assertErrorContents(err, tc.expErr, "QueryParamsCmd error")
			for _, expErr := range tc.expErr {
				s.Assert().Contains(string(outBz), expErr, "QueryParamsCmd output with error")
			}
			if tc.exp != nil {
				act := &sanction.QueryParamsResponse{}
				testFunc := func() {
					err = s.clientCtx.Codec.UnmarshalJSON(outBz, act)
				}
				if s.Assert().NotPanics(testFunc, "UnmarshalJSON on output") {
					if s.Assert().NoError(err, "UnmarshalJSON on output") {
						if !s.Assert().Equal(tc.exp, act, "response") && tc.exp != nil && tc.exp.Params != nil && act != nil && act.Params != nil {
							s.Assert().Equal(tc.exp.Params.ImmediateSanctionMinDeposit.String(),
								act.Params.ImmediateSanctionMinDeposit.String(),
								"ImmediateSanctionMinDeposit")
							s.Assert().Equal(tc.exp.Params.ImmediateUnsanctionMinDeposit.String(),
								act.Params.ImmediateUnsanctionMinDeposit.String(),
								"ImmediateUnsanctionMinDeposit")
						}
					}
				}
			}
		})
	}
}
