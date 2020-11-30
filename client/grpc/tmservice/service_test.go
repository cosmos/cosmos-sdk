package tmservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	queryClient qtypes.ServiceClient
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

	s.queryClient = qtypes.NewServiceClient(s.network.Validators[0].ClientCtx)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s IntegrationTestSuite) TestQueryNodeInfo() {
	val := s.network.Validators[0]

	res, err := s.queryClient.GetNodeInfo(context.Background(), &qtypes.GetNodeInfoRequest{})
	s.Require().NoError(err)
	s.Require().Equal(res.ApplicationVersion.AppName, version.NewInfo().AppName)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/node_info", val.APIAddress))
	s.Require().NoError(err)
	var getInfoRes qtypes.GetNodeInfoResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &getInfoRes))
	s.Require().Equal(getInfoRes.ApplicationVersion.AppName, version.NewInfo().AppName)
}

func (s IntegrationTestSuite) TestQuerySyncing() {
	val := s.network.Validators[0]

	_, err := s.queryClient.GetSyncing(context.Background(), &qtypes.GetSyncingRequest{})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/syncing", val.APIAddress))
	s.Require().NoError(err)
	var syncingRes qtypes.GetSyncingResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &syncingRes))
}

func (s IntegrationTestSuite) TestQueryLatestBlock() {
	val := s.network.Validators[0]

	_, err := s.queryClient.GetLatestBlock(context.Background(), &qtypes.GetLatestBlockRequest{})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", val.APIAddress))
	s.Require().NoError(err)
	var blockInfoRes qtypes.GetLatestBlockResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &blockInfoRes))
}

func (s IntegrationTestSuite) TestQueryBlockByHeight() {
	val := s.network.Validators[0]
	_, err := s.queryClient.GetBlockByHeight(context.Background(), &qtypes.GetBlockByHeightRequest{Height: 1})
	s.Require().NoError(err)

	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/%d", val.APIAddress, 1))
	s.Require().NoError(err)
	var blockInfoRes qtypes.GetBlockByHeightResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &blockInfoRes))
}

func (s IntegrationTestSuite) TestQueryLatestValidatorSet() {
	val := s.network.Validators[0]

	// nil pagination
	_, err := s.queryClient.GetLatestValidatorSet(context.Background(), &qtypes.GetLatestValidatorSetRequest{
		Pagination: nil,
	})
	s.Require().NoError(err)

	//with pagination
	_, err = s.queryClient.GetLatestValidatorSet(context.Background(), &qtypes.GetLatestValidatorSetRequest{Pagination: &qtypes.PageRequest{
		Offset: 0,
		Limit:  10,
	}})
	s.Require().NoError(err)

	// rest request without pagination
	_, err = rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validators/latest", val.APIAddress))
	s.Require().NoError(err)

	// rest request with pagination
	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validators/latest?pagination.offset=%d&pagination.limit=%d", val.APIAddress, 0, 1))
	s.Require().NoError(err)
	var validatorSetRes qtypes.GetLatestValidatorSetResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &validatorSetRes))
}

func (s IntegrationTestSuite) TestQueryValidatorSetByHeight() {
	val := s.network.Validators[0]

	// nil pagination
	_, err := s.queryClient.GetValidatorSetByHeight(context.Background(), &qtypes.GetValidatorSetByHeightRequest{
		Height:     1,
		Pagination: nil,
	})
	s.Require().NoError(err)

	_, err = s.queryClient.GetValidatorSetByHeight(context.Background(), &qtypes.GetValidatorSetByHeightRequest{
		Height: 1,
		Pagination: &qtypes.PageRequest{
			Offset: 0,
			Limit:  10,
		}})
	s.Require().NoError(err)

	// no pagination rest
	_, err = rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validators/%d", val.APIAddress, 1))
	s.Require().NoError(err)

	// rest query with pagination
	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/validators/%d?pagination.offset=%d&pagination.limit=%d", val.APIAddress, 1, 0, 1))
	var validatorSetRes qtypes.GetValidatorSetByHeightResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &validatorSetRes))
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
