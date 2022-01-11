//go:build norace
// +build norace

package rest_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *IntegrationTestSuite) TestLegacyGetValidators() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{
			"old status should show error message",
			fmt.Sprintf("%s/staking/validators?status=bonded", baseURL),
			true, "cosmos sdk v0.40 introduces a breaking change on this endpoint: instead of" +
				" querying using `?status=bonded`, please use `status=BOND_STATUS_BONDED`. For more" +
				" info, please see our REST endpoint migration guide at https://docs.cosmos.network/master/migrations/rest.html",
		},
		{
			"new status should work",
			fmt.Sprintf("%s/staking/validators?status=BOND_STATUS_BONDED", baseURL),
			false, "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			respJSON, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			if tc.expErr {
				var errResp rest.ErrorResponse
				s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(respJSON, &errResp))

				s.Require().Equal(errResp.Error, tc.expErrMsg)
			} else {
				var resp = rest.ResponseWithHeight{}
				err = val.ClientCtx.LegacyAmino.UnmarshalJSON(respJSON, &resp)
				s.Require().NoError(err)

				// Check result is not empty.
				var validators []types.Validator
				s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(resp.Result, &validators))
				s.Require().Greater(len(validators), 0)
				// While we're at it, also check that the consensus_pubkey is
				// an Any, and not bech32 anymore.
				s.Require().Contains(string(resp.Result), "\"consensus_pubkey\": {\n      \"type\": \"tendermint/PubKeyEd25519\",")
			}
		})
	}
}
