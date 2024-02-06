package tmservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/rest"
	"github.com/cosmos/cosmos-sdk/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg         network.Config
	network     *network.Network
	queryClient tmservice.ServiceClient
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.queryClient = tmservice.NewServiceClient(s.network.Validators[0].ClientCtx)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryNodeInfo() {
	val := s.network.Validators[0]

	res, err := s.queryClient.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.ApplicationVersion.AppName, version.NewInfo().AppName)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/node_info", val.APIAddress))
	s.Require().NoError(err)
	var getInfoRes tmservice.GetNodeInfoResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(restRes, &getInfoRes))
	s.Require().Equal(getInfoRes.ApplicationVersion.AppName, version.NewInfo().AppName)
}

func (s *IntegrationTestSuite) TestQuerySyncing() {
	val := s.network.Validators[0]

	_, err := s.queryClient.GetSyncing(context.Background(), &tmservice.GetSyncingRequest{})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/syncing", val.APIAddress))
	s.Require().NoError(err)
	var syncingRes tmservice.GetSyncingResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(restRes, &syncingRes))
}

func (s *IntegrationTestSuite) TestQueryLatestBlock() {
	val := s.network.Validators[0]

	_, err := s.queryClient.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", val.APIAddress))
	s.Require().NoError(err)
	var blockInfoRes tmservice.GetLatestBlockResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(restRes, &blockInfoRes))
	s.Require().Equal(types.ConsAddress(blockInfoRes.Block.Header.ProposerAddress).String(), blockInfoRes.SdkBlock.Header.ProposerAddress)
	s.Require().Contains(blockInfoRes.SdkBlock.Header.ProposerAddress, "cosmosvalcons")
}

func (s *IntegrationTestSuite) TestQueryBlockByHeight() {
	val := s.network.Validators[0]
	_, err := s.queryClient.GetBlockByHeight(context.Background(), &tmservice.GetBlockByHeightRequest{Height: 1})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/%d", val.APIAddress, 1))
	s.Require().NoError(err)
	var blockInfoRes tmservice.GetBlockByHeightResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(restRes, &blockInfoRes))
	s.Require().Contains(blockInfoRes.SdkBlock.Header.ProposerAddress, "cosmosvalcons")
}

func (s *IntegrationTestSuite) TestQueryLatestValidatorSet() {
	val := s.network.Validators[0]

	// nil pagination
	res, err := s.queryClient.GetLatestValidatorSet(context.Background(), &tmservice.GetLatestValidatorSetRequest{
		Pagination: nil,
	})
	s.Require().NoError(err)
	s.Require().Equal(1, len(res.Validators))
	content, ok := res.Validators[0].PubKey.GetCachedValue().(cryptotypes.PubKey)
	s.Require().Equal(true, ok)
	s.Require().Equal(content, val.PubKey)

	// with pagination
	_, err = s.queryClient.GetLatestValidatorSet(context.Background(), &tmservice.GetLatestValidatorSetRequest{Pagination: &qtypes.PageRequest{
		Offset: 0,
		Limit:  10,
	}})
	s.Require().NoError(err)

	// rest request without pagination
	_, err = rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest", val.APIAddress))
	s.Require().NoError(err)

	// rest request with pagination
	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=%d&pagination.limit=%d", val.APIAddress, 0, 1))
	s.Require().NoError(err)
	var validatorSetRes tmservice.GetLatestValidatorSetResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(restRes, &validatorSetRes))
	s.Require().Equal(1, len(validatorSetRes.Validators))
	anyPub, err := codectypes.NewAnyWithValue(val.PubKey)
	s.Require().NoError(err)
	s.Require().Equal(validatorSetRes.Validators[0].PubKey, anyPub)
}

func (s *IntegrationTestSuite) TestLatestValidatorSet_GRPC() {
	vals := s.network.Validators
	testCases := []struct {
		name      string
		req       *tmservice.GetLatestValidatorSetRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "cannot be nil"},
		{"no pagination", &tmservice.GetLatestValidatorSetRequest{}, false, ""},
		{"with pagination", &tmservice.GetLatestValidatorSetRequest{Pagination: &qtypes.PageRequest{Offset: 0, Limit: uint64(len(vals))}}, false, ""},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			grpcRes, err := s.queryClient.GetLatestValidatorSet(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().Len(grpcRes.Validators, len(vals))
				s.Require().Equal(grpcRes.Pagination.Total, uint64(len(vals)))
				content, ok := grpcRes.Validators[0].PubKey.GetCachedValue().(cryptotypes.PubKey)
				s.Require().Equal(true, ok)
				s.Require().Equal(content, vals[0].PubKey)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestLatestValidatorSet_GRPCGateway() {
	vals := s.network.Validators
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{"no pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest", vals[0].APIAddress), false, ""},
		{"pagination invalid fields", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=-1&pagination.limit=-2", vals[0].APIAddress), true, "strconv.ParseUint"},
		{"with pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=0&pagination.limit=2", vals[0].APIAddress), false, ""},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			res, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tmservice.GetLatestValidatorSetResponse
				err = vals[0].ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().Equal(uint64(len(vals)), result.Pagination.Total)
				anyPub, err := codectypes.NewAnyWithValue(vals[0].PubKey)
				s.Require().NoError(err)
				s.Require().Equal(result.Validators[0].PubKey, anyPub)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestValidatorSetByHeight_GRPC() {
	vals := s.network.Validators
	testCases := []struct {
		name      string
		req       *tmservice.GetValidatorSetByHeightRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tmservice.GetValidatorSetByHeightRequest{}, true, "height must be greater than 0"},
		{"no pagination", &tmservice.GetValidatorSetByHeightRequest{Height: 1}, false, ""},
		{"with pagination", &tmservice.GetValidatorSetByHeightRequest{Height: 1, Pagination: &qtypes.PageRequest{Offset: 0, Limit: 1}}, false, ""},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			grpcRes, err := s.queryClient.GetValidatorSetByHeight(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().Len(grpcRes.Validators, len(vals))
				s.Require().Equal(grpcRes.Pagination.Total, uint64(len(vals)))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestValidatorSetByHeight_GRPCGateway() {
	vals := s.network.Validators
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{"invalid height", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d", vals[0].APIAddress, -1), true, "height must be greater than 0"},
		{"no pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d", vals[0].APIAddress, 1), false, ""},
		{"pagination invalid fields", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d?pagination.offset=-1&pagination.limit=-2", vals[0].APIAddress, 1), true, "strconv.ParseUint"},
		{"with pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d?pagination.offset=0&pagination.limit=2", vals[0].APIAddress, 1), false, ""},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			res, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tmservice.GetValidatorSetByHeightResponse
				err = vals[0].ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().Equal(uint64(len(vals)), result.Pagination.Total)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestABCIQuery() {
	testCases := []struct {
		name         string
		req          *tmservice.ABCIQueryRequest
		expectErr    bool
		expectedCode uint32
		validQuery   bool
	}{
		{
			name: "valid request with proof",
			req: &tmservice.ABCIQueryRequest{
				Path:  "/store/gov/key",
				Data:  []byte{0x03},
				Prove: true,
			},
			validQuery: true,
		},
		{
			name: "valid request without proof",
			req: &tmservice.ABCIQueryRequest{
				Path:  "/store/gov/key",
				Data:  []byte{0x03},
				Prove: false,
			},
			validQuery: true,
		},
		{
			name: "request with invalid path",
			req: &tmservice.ABCIQueryRequest{
				Path: "/foo/bar",
				Data: []byte{0x03},
			},
			expectErr: true,
		},
		{
			name: "request with invalid path recursive",
			req: &tmservice.ABCIQueryRequest{
				Path: "/cosmos.base.tendermint.v1beta1.Service/ABCIQuery",
				Data: s.cfg.Codec.MustMarshal(&tmservice.ABCIQueryRequest{
					Path: "/cosmos.base.tendermint.v1beta1.Service/ABCIQuery",
				}),
			},
			expectErr: true,
		},
		{
			name: "request with invalid broadcast tx path",
			req: &tmservice.ABCIQueryRequest{
				Path: "/cosmos.tx.v1beta1.Service/BroadcastTx",
				Data: []byte{0x00},
			},
			expectErr: true,
		},
		{
			name: "request with invalid data",
			req: &tmservice.ABCIQueryRequest{
				Path: "/store/gov/key",
				Data: []byte{0x0044, 0x00},
			},
			validQuery: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			res, err := s.queryClient.ABCIQuery(context.Background(), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Nil(res)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Equal(res.Code, tc.expectedCode)
			}

			if tc.validQuery {
				s.Require().Greater(res.Height, int64(0))
				s.Require().Greater(len(res.Key), 0, "expected non-empty key")
				s.Require().Greater(len(res.Value), 0, "expected non-empty value")
			}

			if tc.req.Prove {
				s.Require().Greater(len(res.ProofOps.Ops), 0, "expected proofs")
			}
		})
	}
}
