package feegrant

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"
)

func (s *E2ETestSuite) TestGrantGRPCAllowance() {
	val := s.network.Validators[0]
	s.createGrant(s.addedGranter, s.addedGrantee)
	testCases := []struct {
		name string
		args struct {
			Granter string
			Grantee string
		}
		expectErr bool
		errorMsg  string
		msg       string
	}{
		{
			name: "correct grant",
			args: struct {
				Granter string
				Grantee string
			}{
				Granter: s.addedGranter.String(),
				Grantee: s.addedGrantee.String(),
			},
			expectErr: false,
			errorMsg:  "",
			msg:       "\"allowance\":{\"@type\":\"/cosmos.feegrant.v1beta1.BasicAllowance\",\"spend_limit\":[{\"denom\":\"stake\",\"amount\":\"100\"}]",
		},
		{
			name: "invalid grantee",
			args: struct {
				Granter string
				Grantee string
			}{
				Granter: s.addedGranter.String(),
				Grantee: "incorrect_grantee",
			},
			expectErr: true,
			errorMsg:  "decoding bech32 failed",
			msg:       "",
		},
		{
			name: "non existent grant",
			args: struct {
				Granter string
				Grantee string
			}{
				Granter: "cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
				Grantee: s.addedGrantee.String(),
			},
			expectErr: true,
			errorMsg:  "fee-grant not found",
			msg:       "",
		},
	}

	grantAllowneceURL := val.APIAddress + "/cosmos/feegrant/v1beta1/allowance/%s/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(grantAllowneceURL, tc.args.Granter, tc.args.Grantee)
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				s.Require().Contains(string(resp), tc.msg)
			}
		})
	}
}

func (s *E2ETestSuite) TestGrantGRPCAllowanceByGrantee() {
	val := s.network.Validators[0]
	s.createGrant(s.addedGranter, s.addedGrantee)
	testCases := []struct {
		name string
		args struct {
			Grantee string
		}
		expectErr bool
		errorMsg  string
		msg       string
	}{
		{
			name: "correct grpc request",
			args: struct {
				Grantee string
			}{
				Grantee: s.addedGrantee.String(),
			},
			expectErr: false,
			errorMsg:  "",
			msg:       "\"allowance\":{\"@type\":\"/cosmos.feegrant.v1beta1.BasicAllowance\",\"spend_limit\":[{\"denom\":\"stake\",\"amount\":\"100\"}]",
		},
		{
			name: "invalid grantee",
			args: struct {
				Grantee string
			}{
				Grantee: "incorrect_grantee",
			},
			expectErr: true,
			errorMsg:  "decoding bech32 failed",
			msg:       "",
		},
		{
			name: "non existent grant",
			args: struct {
				Grantee string
			}{
				Grantee: "cosmos1h6wddg9x0zsusswhchkfhwwtkdg62fehs5ees4",
			},
			expectErr: false,
			errorMsg:  "",
			msg:       "{\"allowances\":[],\"pagination\":{\"next_key\":null,\"total\":\"0\"}}",
		},
	}

	grantAllowneceURL := val.APIAddress + "/cosmos/feegrant/v1beta1/allowances/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(grantAllowneceURL, tc.args.Grantee)
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				s.Require().Contains(string(resp), tc.msg)
			}
		})
	}
}

func (s *E2ETestSuite) TestGrantGRPCAllowanceByGranter() {
	val := s.network.Validators[0]
	s.createGrant(s.addedGranter, s.addedGrantee)
	testCases := []struct {
		name string
		args struct {
			Granter string
		}
		expectErr bool
		errorMsg  string
		msg       string
	}{
		{
			name: "correct grpc request",
			args: struct {
				Granter string
			}{
				Granter: s.addedGranter.String(),
			},
			expectErr: false,
			errorMsg:  "",
			msg:       "\"allowance\":{\"@type\":\"/cosmos.feegrant.v1beta1.BasicAllowance\",\"spend_limit\":[{\"denom\":\"stake\",\"amount\":\"100\"}]",
		},
		{
			name: "invalid grantee",
			args: struct {
				Granter string
			}{
				Granter: "incorrect_grantee",
			},
			expectErr: true,
			errorMsg:  "decoding bech32 failed",
			msg:       "",
		},
		{
			name: "non existent grant",
			args: struct {
				Granter string
			}{
				Granter: "cosmos1h6wddg9x0zsusswhchkfhwwtkdg62fehs5ees4",
			},
			expectErr: false,
			errorMsg:  "{\"allowances\":[],\"pagination\":{\"next_key\":null,\"total\":\"0\"}}",
			msg:       "",
		},
	}

	grantAllowneceURL := val.APIAddress + "/cosmos/feegrant/v1beta1/issued/%s"
	for _, tc := range testCases {
		uri := fmt.Sprintf(grantAllowneceURL, tc.args.Granter)
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(uri)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				s.Require().NoError(err)
				s.Require().Contains(string(resp), tc.msg)
			}
		})
	}
}
