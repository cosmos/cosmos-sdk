//go:build e2e
// +build e2e

package cmtservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	_ "cosmossdk.io/x/gov"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

type E2ETestSuite struct {
	suite.Suite

	cfg         network.Config
	network     network.NetworkI
	queryClient cmtservice.ServiceClient
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	s.queryClient = cmtservice.NewServiceClient(s.network.GetValidators()[0].GetClientCtx())
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestQueryNodeInfo() {
	val := s.network.GetValidators()[0]

	res, err := s.queryClient.GetNodeInfo(context.Background(), &cmtservice.GetNodeInfoRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.ApplicationVersion.AppName, version.NewInfo().AppName)

	restRes, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/node_info", val.GetAPIAddress()))
	s.Require().NoError(err)
	var getInfoRes cmtservice.GetNodeInfoResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(restRes, &getInfoRes))
	s.Require().Equal(getInfoRes.ApplicationVersion.AppName, version.NewInfo().AppName)
}

func (s *E2ETestSuite) TestQuerySyncing() {
	val := s.network.GetValidators()[0]

	_, err := s.queryClient.GetSyncing(context.Background(), &cmtservice.GetSyncingRequest{})
	s.Require().NoError(err)

	restRes, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/syncing", val.GetAPIAddress()))
	s.Require().NoError(err)
	var syncingRes cmtservice.GetSyncingResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(restRes, &syncingRes))
}

func (s *E2ETestSuite) TestQueryLatestBlock() {
	val := s.network.GetValidators()[0]

	_, err := s.queryClient.GetLatestBlock(context.Background(), &cmtservice.GetLatestBlockRequest{})
	s.Require().NoError(err)

	restRes, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", val.GetAPIAddress()))
	s.Require().NoError(err)
	var blockInfoRes cmtservice.GetLatestBlockResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(restRes, &blockInfoRes))
	s.Require().Contains(blockInfoRes.SdkBlock.Header.ProposerAddress, "cosmosvalcons")
}

func (s *E2ETestSuite) TestQueryBlockByHeight() {
	val := s.network.GetValidators()[0]
	_, err := s.queryClient.GetBlockByHeight(context.Background(), &cmtservice.GetBlockByHeightRequest{Height: 1})
	s.Require().NoError(err)

	restRes, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/%d", val.GetAPIAddress(), 1))
	s.Require().NoError(err)
	var blockInfoRes cmtservice.GetBlockByHeightResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(restRes, &blockInfoRes))
	s.Require().Contains(blockInfoRes.SdkBlock.Header.ProposerAddress, "cosmosvalcons")
}

func (s *E2ETestSuite) TestQueryLatestValidatorSet() {
	val := s.network.GetValidators()[0]

	// nil pagination
	res, err := s.queryClient.GetLatestValidatorSet(context.Background(), &cmtservice.GetLatestValidatorSetRequest{
		Pagination: nil,
	})
	s.Require().NoError(err)
	s.Require().Equal(1, len(res.Validators))
	content, ok := res.Validators[0].PubKey.GetCachedValue().(cryptotypes.PubKey)
	s.Require().Equal(true, ok)
	s.Require().Equal(content, val.GetPubKey())

	// with pagination
	_, err = s.queryClient.GetLatestValidatorSet(context.Background(), &cmtservice.GetLatestValidatorSetRequest{Pagination: &qtypes.PageRequest{
		Offset: 0,
		Limit:  10,
	}})
	s.Require().NoError(err)

	// rest request without pagination
	_, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest", val.GetAPIAddress()))
	s.Require().NoError(err)

	// rest request with pagination
	restRes, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=%d&pagination.limit=%d", val.GetAPIAddress(), 0, 1))
	s.Require().NoError(err)
	var validatorSetRes cmtservice.GetLatestValidatorSetResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(restRes, &validatorSetRes))
	s.Require().Equal(1, len(validatorSetRes.Validators))
	anyPub, err := codectypes.NewAnyWithValue(val.GetPubKey())
	s.Require().NoError(err)
	s.Require().Equal(validatorSetRes.Validators[0].PubKey, anyPub)
}

func (s *E2ETestSuite) TestLatestValidatorSet_GRPC() {
	vals := s.network.GetValidators()
	testCases := []struct {
		name      string
		req       *cmtservice.GetLatestValidatorSetRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "cannot be nil"},
		{"no pagination", &cmtservice.GetLatestValidatorSetRequest{}, false, ""},
		{"with pagination", &cmtservice.GetLatestValidatorSetRequest{Pagination: &qtypes.PageRequest{Offset: 0, Limit: uint64(len(vals))}}, false, ""},
	}
	for _, tc := range testCases {
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
				s.Require().Equal(content, vals[0].GetPubKey())
			}
		})
	}
}

func (s *E2ETestSuite) TestLatestValidatorSet_GRPCGateway() {
	vals := s.network.GetValidators()
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{"no pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest", vals[0].GetAPIAddress()), false, ""},
		{"pagination invalid fields", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=-1&pagination.limit=-2", vals[0].GetAPIAddress()), true, "strconv.ParseUint"},
		{"with pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/latest?pagination.offset=0&pagination.limit=2", vals[0].GetAPIAddress()), false, ""},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result cmtservice.GetLatestValidatorSetResponse
				err = vals[0].GetClientCtx().Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().Equal(uint64(len(vals)), result.Pagination.Total)
				anyPub, err := codectypes.NewAnyWithValue(vals[0].GetPubKey())
				s.Require().NoError(err)
				s.Require().Equal(result.Validators[0].PubKey, anyPub)
			}
		})
	}
}

func (s *E2ETestSuite) TestValidatorSetByHeight_GRPC() {
	vals := s.network.GetValidators()
	testCases := []struct {
		name      string
		req       *cmtservice.GetValidatorSetByHeightRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &cmtservice.GetValidatorSetByHeightRequest{}, true, "height must be greater than 0"},
		{"no pagination", &cmtservice.GetValidatorSetByHeightRequest{Height: 1}, false, ""},
		{"with pagination", &cmtservice.GetValidatorSetByHeightRequest{Height: 1, Pagination: &qtypes.PageRequest{Offset: 0, Limit: 1}}, false, ""},
	}
	for _, tc := range testCases {
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

func (s *E2ETestSuite) TestValidatorSetByHeight_GRPCGateway() {
	vals := s.network.GetValidators()
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{"invalid height", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d", vals[0].GetAPIAddress(), -1), true, "height must be greater than 0"},
		{"no pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d", vals[0].GetAPIAddress(), 1), false, ""},
		{"pagination invalid fields", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d?pagination.offset=-1&pagination.limit=-2", vals[0].GetAPIAddress(), 1), true, "strconv.ParseUint"},
		{"with pagination", fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d?pagination.offset=0&pagination.limit=2", vals[0].GetAPIAddress(), 1), false, ""},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result cmtservice.GetValidatorSetByHeightResponse
				err = vals[0].GetClientCtx().Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().Equal(uint64(len(vals)), result.Pagination.Total)
			}
		})
	}
}

func (s *E2ETestSuite) TestABCIQuery() {
	testCases := []struct {
		name         string
		req          *cmtservice.ABCIQueryRequest
		expectErr    bool
		expectedCode uint32
		validQuery   bool
	}{
		{
			name: "valid request with proof",
			req: &cmtservice.ABCIQueryRequest{
				Path:  "/store/gov/key",
				Data:  []byte{0x03},
				Prove: true,
			},
			validQuery: true,
		},
		{
			name: "valid request without proof",
			req: &cmtservice.ABCIQueryRequest{
				Path:  "/store/gov/key",
				Data:  []byte{0x03},
				Prove: false,
			},
			validQuery: true,
		},
		{
			name: "request with invalid path",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/foo/bar",
				Data: []byte{0x03},
			},
			expectErr: true,
		},
		{
			name: "request with invalid path recursive",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/cosmos.base.tendermint.v1beta1.Service/ABCIQuery",
				Data: s.cfg.Codec.MustMarshal(&cmtservice.ABCIQueryRequest{
					Path: "/cosmos.base.tendermint.v1beta1.Service/ABCIQuery",
				}),
			},
			expectErr: true,
		},
		{
			name: "request with invalid broadcast tx path",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/cosmos.tx.v1beta1.Service/BroadcastTx",
				Data: []byte{0x00},
			},
			expectErr: true,
		},
		{
			name: "request with invalid data",
			req: &cmtservice.ABCIQueryRequest{
				Path: "/store/gov/key",
				Data: []byte{0x0044, 0x00},
			},
			validQuery: false,
		},
	}

	for _, tc := range testCases {
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
