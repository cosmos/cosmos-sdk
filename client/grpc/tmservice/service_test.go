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
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	queryClient tmservice.ServiceClient
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.queryClient = tmservice.NewServiceClient(s.network.Validators[0].ClientCtx)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s IntegrationTestSuite) TestQueryNodeInfo() {
	val := s.network.Validators[0]

	res, err := s.queryClient.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.ApplicationVersion.AppName, version.NewInfo().AppName)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/node_info", val.APIAddress))
	s.Require().NoError(err)
	var getInfoRes tmservice.GetNodeInfoResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &getInfoRes))
	s.Require().Equal(getInfoRes.ApplicationVersion.AppName, version.NewInfo().AppName)
}

func (s IntegrationTestSuite) TestQuerySyncing() {
	val := s.network.Validators[0]

	_, err := s.queryClient.GetSyncing(context.Background(), &tmservice.GetSyncingRequest{})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/syncing", val.APIAddress))
	s.Require().NoError(err)
	var syncingRes tmservice.GetSyncingResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &syncingRes))
}

func (s IntegrationTestSuite) TestQueryLatestBlock() {
	val := s.network.Validators[0]

	_, err := s.queryClient.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", val.APIAddress))
	s.Require().NoError(err)
	var blockInfoRes tmservice.GetLatestBlockResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &blockInfoRes))
}

func (s IntegrationTestSuite) TestQueryBlockByHeight() {
	val := s.network.Validators[0]
	_, err := s.queryClient.GetBlockByHeight(context.Background(), &tmservice.GetBlockByHeightRequest{Height: 1})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/%d", val.APIAddress, 1))
	s.Require().NoError(err)
	var blockInfoRes tmservice.GetBlockByHeightResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &blockInfoRes))
}

func (s IntegrationTestSuite) TestQueryLatestValidatorSet() {
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

	//with pagination
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
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &validatorSetRes))
	s.Require().Equal(1, len(validatorSetRes.Validators))
	anyPub, err := codectypes.NewAnyWithValue(val.PubKey)
	s.Require().NoError(err)
	s.Require().Equal(validatorSetRes.Validators[0].PubKey, anyPub)
}

func (s IntegrationTestSuite) TestQueryValidatorSetByHeight() {
	val := s.network.Validators[0]

	// nil pagination
	_, err := s.queryClient.GetValidatorSetByHeight(context.Background(), &tmservice.GetValidatorSetByHeightRequest{
		Height:     1,
		Pagination: nil,
	})
	s.Require().NoError(err)

	_, err = s.queryClient.GetValidatorSetByHeight(context.Background(), &tmservice.GetValidatorSetByHeightRequest{
		Height: 1,
		Pagination: &qtypes.PageRequest{
			Offset: 0,
			Limit:  10,
		}})
	s.Require().NoError(err)

	// no pagination rest
	_, err = rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d", val.APIAddress, 1))
	s.Require().NoError(err)

	// rest query with pagination
	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validatorsets/%d?pagination.offset=%d&pagination.limit=%d", val.APIAddress, 1, 0, 1))
	var validatorSetRes tmservice.GetValidatorSetByHeightResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &validatorSetRes))
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
