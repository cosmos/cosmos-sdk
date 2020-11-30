package rest_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	msgauthtestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
	types "github.com/cosmos/cosmos-sdk/x/authz/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
	grantee sdk.AccAddress
}

var typeMsgSend = types.SendAuthorization{}.MethodName()

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()

	cfg.NumValidators = 1
	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	val := s.network.Validators[0]
	// Create new account in the keyring.
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)
	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	// Send some funds to the new account.
	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	// grant authorization
	_, err = msgauthtestutil.MsgGrantAuthorizationExec(val.ClientCtx, val.Address.String(), newAddr.String(), typeMsgSend, "100stake")
	s.Require().NoError(err)

	_, err = msgauthtestutil.MsgGrantAuthorizationExec(val.ClientCtx, val.Address.String(), newAddr.String(), "GericAuthorization", "")
	s.Require().NoError(err)

	s.grantee = newAddr
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryAuthorizationGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		respMsg   proto.Message
		expected  proto.Message
	}{
		{
			"fail invalid granter address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?msg_type=%s", baseURL, "invalid_granter", s.grantee.String(), typeMsgSend),
			true,
			&types.QueryAuthorizationResponse{},
			&types.QueryAuthorizationResponse{},
		},
		{
			"fail invalid grantee address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?msg_type=%s", baseURL, val.Address.String(), "invalid_grantee", typeMsgSend),
			true,
			&types.QueryAuthorizationResponse{},
			&types.QueryAuthorizationResponse{},
		},
		{
			"fail invalid msg-type",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?msg_type=%s", baseURL, val.Address.String(), s.grantee.String(), "invalidMsg"),
			true,
			&types.QueryAuthorizationResponse{},
			&types.QueryAuthorizationResponse{},
		},
		{
			"valid query",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?msg_type=%s", baseURL, val.Address.String(), s.grantee.String(), typeMsgSend),
			false,
			&types.QueryAuthorizationResponse{},
			&types.QueryAuthorizationResponse{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respMsg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryAuthorizationsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		respMsg   proto.Message
		expected  proto.Message
	}{
		{
			"fail invalid granter address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants", baseURL, "invalid_granter", s.grantee.String()),
			true,
			&types.QueryAuthorizationsResponse{},
			&types.QueryAuthorizationsResponse{},
		},
		{
			"fail invalid grantee address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants", baseURL, val.Address.String(), "invalid_grantee"),
			true,
			&types.QueryAuthorizationsResponse{},
			&types.QueryAuthorizationsResponse{},
		},
		{
			"valid query",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants?pagination.limit=1&pagination.count_total=true", baseURL, val.Address.String(), s.grantee.String()),
			false,
			&types.QueryAuthorizationsResponse{},
			&types.QueryAuthorizationsResponse{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respMsg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
