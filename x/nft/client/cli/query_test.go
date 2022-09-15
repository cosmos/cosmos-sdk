package cli_test

import (
	"fmt"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/cosmos/cosmos-sdk/x/nft/client/cli"
)

func (s *CLITestSuite) TestQueryClass() {
	testCases := []struct {
		name string
		args struct {
			ClassID string
		}
		expectErr bool
	}{
		{
			name: "valid case",
			args: struct {
				ClassID string
			}{
				ClassID: testClassID,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryClass()
			var args []string
			args = append(args, tc.args.ClassID)
			args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryClassResponse
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestQueryClasses() {
	testCases := []struct {
		name      string
		expectErr bool
	}{
		{
			name:      "no params",
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryClasses()
			var args []string
			args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryClassesResponse
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestQueryNFT() {
	testCases := []struct {
		name string
		args struct {
			ClassID string
			ID      string
		}
		expectErr bool
	}{
		{
			name: "valid case",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: testClassID,
				ID:      testID,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryNFT()
			var args []string
			args = append(args, tc.args.ClassID)
			args = append(args, tc.args.ID)
			args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTResponse
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestQueryNFTs() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name string
		args struct {
			ClassID string
			Owner   string
		}
		expectErr bool
	}{
		{
			name: "empty class id and owner",
			args: struct {
				ClassID string
				Owner   string
			}{},
			expectErr: true,
		},
		{
			name: "valid case",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: testClassID,
				Owner:   accounts[0].Address.String(),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryNFTs()
			var args []string
			args = append(args, fmt.Sprintf("--%s=%s", cli.FlagClassID, tc.args.ClassID))
			args = append(args, fmt.Sprintf("--%s=%s", cli.FlagOwner, tc.args.Owner))
			args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryNFTsResponse
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestQueryOwner() {
	testCases := []struct {
		name string
		args struct {
			ClassID string
			ID      string
		}
		expectErr bool
	}{
		{
			name: "valid case",
			args: struct {
				ClassID string
				ID      string
			}{
				ClassID: testClassID,
				ID:      testID,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryOwner()
			var args []string
			args = append(args, tc.args.ClassID)
			args = append(args, tc.args.ID)
			args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryOwnerResponse
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestQueryBalance() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name string
		args struct {
			ClassID string
			Owner   string
		}
		expectErr bool
	}{
		{
			name: "valid case",
			args: struct {
				ClassID string
				Owner   string
			}{
				ClassID: testClassID,
				Owner:   accounts[0].Address.String(),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryBalance()
			var args []string
			args = append(args, tc.args.Owner)
			args = append(args, tc.args.ClassID)
			args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QueryBalanceResponse
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestQuerySupply() {
	testCases := []struct {
		name string
		args struct {
			ClassID string
		}
		expectErr bool
	}{
		{
			name: "valid case",
			args: struct {
				ClassID string
			}{
				ClassID: testClassID,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQuerySupply()
			var args []string
			args = append(args, tc.args.ClassID)
			args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var result nft.QuerySupplyResponse
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result)
				s.Require().NoError(err)
			}
		})
	}
}
