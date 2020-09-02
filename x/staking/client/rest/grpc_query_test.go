package rest_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 2

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	unbond, err := sdk.ParseCoin("10stake")
	s.Require().NoError(err)

	val := s.network.Validators[0]
	val2 := s.network.Validators[1]

	// redelegate
	_, err = stakingtestutil.MsgRedelegateExec(val.ClientCtx, val.Address, val.ValAddress, val2.ValAddress, unbond)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// unbonding
	_, err = stakingtestutil.MsgUnbondExec(val.ClientCtx, val.Address, val.ValAddress, unbond)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryValidatorsGRPCHandler() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		error   bool
	}{
		{
			"test query validators gRPC route with invalid status",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=active", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
		},
		{
			"test query validators gRPC route without status query param",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
		},
		{
			"test query validators gRPC route with valid status",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=%s", baseURL, sdk.Bonded.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var valRes types.QueryValidatorsResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &valRes)

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

func (s *IntegrationTestSuite) TestQueryValidatorGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	valAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		error   bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", baseURL, "wrongValidatorAddress"),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", baseURL, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", baseURL, valAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var validator types.QueryValidatorResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &validator)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(validator.Validator)
				s.Require().Equal(s.network.Validators[0].ValAddress, validator.Validator.OperatorAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryValidatorDelegationsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	valAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

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
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryValidatorDelegationsResponse{},
			nil,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations", baseURL, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryValidatorDelegationsResponse{},
			nil,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations", baseURL, valAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryValidatorDelegationsResponse{},
			&types.QueryValidatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDec(100000000), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000))),
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

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResp.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryValidatorUnbondingDelegationsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	valAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

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
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations", baseURL, valAddressBase64),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var ubds types.QueryValidatorUnbondingDelegationsResponse

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &ubds)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, 1)
				s.Require().Equal(ubds.UnbondingResponses[0].ValidatorAddress, val.ValAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryDelegationGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	valAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)
	accAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

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
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, "wrongValAddress", accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"wrong account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, valAddressBase64, "wrongAccAddress"),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, "", accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"with no account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, valAddressBase64, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryDelegationResponse{},
			nil,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", baseURL, valAddressBase64, accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryDelegationResponse{},
			&types.QueryDelegationResponse{
				DelegationResponse: &types.DelegationResponse{
					Delegation: types.Delegation{
						DelegatorAddress: val.Address,
						ValidatorAddress: val.ValAddress,
						Shares:           sdk.NewDec(100000000),
					},
					Balance: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000)),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResp.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryUnbondingDelegationGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	valAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)
	accAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		error   bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, "wrongValAddress", accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"wrong account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, valAddressBase64, "wrongAccAddress"),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, "", accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"with no account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, valAddressBase64, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", baseURL, valAddressBase64, accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "3",
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var ubd types.QueryUnbondingDelegationResponse

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &ubd)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(ubd.Unbond.DelegatorAddress, val.Address)
				s.Require().Equal(ubd.Unbond.ValidatorAddress, val.ValAddress)
				s.Require().Len(ubd.Unbond.Entries, 1)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryDelegatorDelegationsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	delAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

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
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryDelegatorDelegationsResponse{},
			nil,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", baseURL, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryDelegatorDelegationsResponse{},
			nil,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", baseURL, delAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryDelegatorDelegationsResponse{},
			&types.QueryDelegatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDec(100000000), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000))),
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

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedResp.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryDelegatorUnbondingDelegationsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	delAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

	testCases := []struct {
		name       string
		url        string
		headers    map[string]string
		error      bool
		ubdsLength int
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", baseURL, "wrongValAddress"),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
			0,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", baseURL, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
			0,
		},
		{
			"valid request with less height",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", baseURL, delAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
			0,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", baseURL, delAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "3",
			},
			false,
			1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var ubds types.QueryDelegatorUnbondingDelegationsResponse

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &ubds)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, tc.ubdsLength)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryRedelegationsGRPC() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]
	baseURL := val.APIAddress
	delAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)
	srcValBase64 := base64.URLEncoding.EncodeToString(val.Address)
	dstValBase64 := base64.URLEncoding.EncodeToString(val2.Address)

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		error   bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations", baseURL, "wrongValAddress"),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations", baseURL, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations", baseURL, delAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "3",
			},
			false,
		},
		{
			"valid request with src address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations?src_validator_addr=%s", baseURL, delAddressBase64, srcValBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "3",
			},
			false,
		},
		{
			"valid request with dst address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations?dst_validator_addr=%s", baseURL, delAddressBase64, dstValBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "3",
			},
			false,
		},
		{
			"valid request with dst address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/redelegations?src_validator_addr=%s&dst_validator_addr=%s", baseURL, delAddressBase64, srcValBase64, dstValBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "3",
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)
			var redelegations types.QueryRedelegationsResponse

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &redelegations)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryDelegatorValidatorsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	delAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		error   bool
	}{
		{
			"wrong delegator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators", baseURL, "wrongDelAddress"),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"with no delegator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators", baseURL, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators", baseURL, delAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var validators types.QueryDelegatorValidatorsResponse

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &validators)

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

func (s *IntegrationTestSuite) TestQueryDelegatorValidatorGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	valAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)
	accAddressBase64 := base64.URLEncoding.EncodeToString(val.Address)

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		error   bool
	}{
		{
			"wrong validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, "wrongValAddress", accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"wrong account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, valAddressBase64, "wrongAccAddress"),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"with no validator address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, "", accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"with no account address",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, valAddressBase64, ""),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			true,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/validators/%s", baseURL, valAddressBase64, accAddressBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var validator types.QueryDelegatorValidatorResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &validator)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(validator)
				s.Require().Equal(validator.Validator.OperatorAddress, val.ValAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryHistoricalInfoGRPC() {
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

			var historical_info types.QueryHistoricalInfoResponse

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &historical_info)

			if tc.error {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(historical_info)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryParamsGRPC() {
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
			s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}

func (s *IntegrationTestSuite) TestQueryPoolGRPC() {
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
			fmt.Sprintf("%s/cosmos/staking/v1beta1/pool", baseURL),
			&types.QueryPoolResponse{},
			&types.QueryPoolResponse{
				Pool: types.Pool{
					NotBondedTokens: sdk.NewInt(10),
					BondedTokens:    sdk.NewInt(199999990),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := rest.GetRequest(tc.url)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
