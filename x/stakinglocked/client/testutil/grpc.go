package testutil

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/rest"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *IntegrationTestSuite) TestGRPCQueryValidatorsHandler() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"test query validators gRPC route with invalid status",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=active", baseURL),
			true,
		},
		{
			"test query validators gRPC route without status query param",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", baseURL),
			false,
		},
		{
			"test query validators gRPC route with valid status",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=%s", baseURL, types.Bonded.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var valRes types.QueryValidatorsResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &valRes)

			if tc.error {
				s.Require().Error(err)
				s.Require().Nil(valRes.Validators)
				s.Require().Equal(0, len(valRes.Validators))
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(valRes.Validators)
				s.Require().Equal(len(s.network.Validators), len(valRes.Validators))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryValidator() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", baseURL, "wrongValidatorAddress"),
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", baseURL, ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", baseURL, val.ValAddress.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var validator types.QueryValidatorResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &validator)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(validator.Validator)
				s.Require().Equal(s.network.Validators[0].ValAddress.String(), validator.Validator.OperatorAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryValidatorDelegations() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name         string
		url          string
		headers      map[string]string
		error        bool
		respType     proto.Message
		expectedResp proto.Message
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations", baseURL, "wrongValAddress"),
			map[string]string{},
			true,
			&types.QueryValidatorDelegationsResponse{},
			nil,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations", baseURL, ""),
			map[string]string{},
			true,
			&types.QueryValidatorDelegationsResponse{},
			nil,
		},
		{
			"valid request(height specific)",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations", baseURL, val.ValAddress.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryValidatorDelegationsResponse{},
			&types.QueryValidatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDecFromInt(cli.DefaultTokens), sdk.NewCoin(sdk.DefaultBondDenom, cli.DefaultTokens)),
				},
				Pagination: &query.PageResponse{Total: 1},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResp.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryValidatorUnbondingDelegations() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations", baseURL, "wrongValAddress"),
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations", baseURL, ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations", baseURL, val.ValAddress.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var ubds types.QueryValidatorUnbondingDelegationsResponse

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &ubds)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, 1)
				s.Require().Equal(ubds.UnbondingResponses[0].ValidatorAddress, val.ValAddress.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryDelegation() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]
	baseURL := val.APIAddress

	testCases := []struct {
		name         string
		url          string
		error        bool
		respType     proto.Message
		expectedResp proto.Message
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, "wrongValAddress", val.Address.String()),
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"wrong account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, val.ValAddress.String(), "wrongAccAddress"),
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, "", val.Address.String()),
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"with no account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, val.ValAddress.String(), ""),
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, val2.ValAddress.String(), val.Address.String()),
			false,
			&types.QueryDelegationResponse{},
			&types.QueryDelegationResponse{
				DelegationResponse: &types.DelegationResponse{
					Delegation: types.Delegation{
						DelegatorAddress: val.Address.String(),
						ValidatorAddress: val2.ValAddress.String(),
						Shares:           sdk.NewDec(10),
					},
					Balance: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)
			s.T().Logf("%s", resp)
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResp.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryUnbondingDelegation() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, "wrongValAddress", val.Address.String()),
			true,
		},
		{
			"wrong account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, val.ValAddress.String(), "wrongAccAddress"),
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, "", val.Address.String()),
			true,
		},
		{
			"with no account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, val.ValAddress.String(), ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, val.ValAddress.String(), val.Address.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var ubd types.QueryUnbondingDelegationResponse

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &ubd)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(ubd.Unbond.DelegatorAddress, val.Address.String())
				s.Require().Equal(ubd.Unbond.ValidatorAddress, val.ValAddress.String())
				s.Require().Len(ubd.Unbond.Entries, 2)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryDelegatorDelegations() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	// Create new account in the keyring for address without delegations.
	k, _, err := val.ClientCtx.Keyring.NewMnemonic("test", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	newAddr, err := k.GetAddress()
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		url          string
		headers      map[string]string
		error        bool
		respType     proto.Message
		expectedResp proto.Message
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", baseURL, "wrongValAddress"),
			map[string]string{},
			true,
			&types.QueryDelegatorDelegationsResponse{},
			nil,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", baseURL, ""),
			map[string]string{},
			true,
			&types.QueryDelegatorDelegationsResponse{},
			nil,
		},
		{
			"valid request (height specific)",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", baseURL, val.Address.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryDelegatorDelegationsResponse{},
			&types.QueryDelegatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDecFromInt(cli.DefaultTokens), sdk.NewCoin(sdk.DefaultBondDenom, cli.DefaultTokens)),
				},
				Pagination: &query.PageResponse{Total: 1},
			},
		},
		{
			"address without delegations",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", baseURL, newAddr.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryDelegatorDelegationsResponse{},
			&types.QueryDelegatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{},
				Pagination:          &query.PageResponse{Total: 0},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResp.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryDelegatorUnbondingDelegations() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name       string
		url        string
		error      bool
		ubdsLength int
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", baseURL, "wrongValAddress"),
			true,
			0,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", baseURL, ""),
			true,
			0,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", baseURL, val.Address.String()),
			false,
			1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var ubds types.QueryDelegatorUnbondingDelegationsResponse

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &ubds)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, tc.ubdsLength)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryRedelegations() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations", baseURL, "wrongValAddress"),
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations", baseURL, ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations", baseURL, val.Address.String()),
			false,
		},
		{
			"valid request with src address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations?src_validator_addr=%s", baseURL, val.Address.String(), val.ValAddress.String()),
			false,
		},
		{
			"valid request with dst address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations?dst_validator_addr=%s", baseURL, val.Address.String(), val.ValAddress.String()),
			false,
		},
		{
			"valid request with dst address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations?src_validator_addr=%s&dst_validator_addr=%s", baseURL, val.Address.String(), val.ValAddress.String(), val2.ValAddress.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)
			var redelegations types.QueryRedelegationsResponse

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &redelegations)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryDelegatorValidators() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"wrong delegator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators", baseURL, "wrongDelAddress"),
			true,
		},
		{
			"with no delegator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators", baseURL, ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators", baseURL, val.Address.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var validators types.QueryDelegatorValidatorsResponse

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &validators)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(validators.Validators, len(s.network.Validators))
				s.Require().Equal(int(validators.Pagination.Total), len(s.network.Validators))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryDelegatorValidator() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"wrong delegator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, "wrongAccAddress", val.ValAddress.String()),
			true,
		},
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, val.Address.String(), "wrongValAddress"),
			true,
		},
		{
			"with empty delegator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, "", val.ValAddress.String()),
			true,
		},
		{
			"with empty validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, val.Address.String(), ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, val.Address.String(), val.ValAddress.String()),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var validator types.QueryDelegatorValidatorResponse
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &validator)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(validator)
				s.Require().Equal(validator.Validator.OperatorAddress, val.ValAddress.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryHistoricalInfo() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name  string
		url   string
		error bool
	}{
		{
			"wrong height",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/historical_info/%s", baseURL, "-1"),
			true,
		},
		{
			"with no height",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/historical_info/%s", baseURL, ""),
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/historical_info/%s", baseURL, "2"),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var historicalInfo types.QueryHistoricalInfoResponse

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, &historicalInfo)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(historicalInfo)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryParams() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request params",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/params", baseURL),
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				Params: types.DefaultParams(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := rest.GetRequest(tc.url)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCQueryPool() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request params",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/pool", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QueryPoolResponse{},
			&types.QueryPoolResponse{
				Pool: types.Pool{
					NotBondedTokens: sdk.NewInt(0),
					BondedTokens:    cli.DefaultTokens.Mul(sdk.NewInt(2)),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}
