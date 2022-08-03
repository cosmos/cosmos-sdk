package keeper_test

import (
	gocontext "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/slashing/testslashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryParams() {
	queryClient := suite.queryClient
	paramsResp, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})

	suite.NoError(err)
	suite.Equal(testslashing.TestParams(), paramsResp.Params)
}

func (suite *KeeperTestSuite) TestGRPCSigningInfo() {
	queryClient := suite.queryClient

	infoResp, err := queryClient.SigningInfo(gocontext.Background(), &types.QuerySigningInfoRequest{ConsAddress: ""})
	suite.Error(err)
	suite.Nil(infoResp)

	consAddr := sdk.ConsAddress(suite.addrDels[0])
	info, found := suite.slashingKeeper.GetValidatorSigningInfo(suite.ctx, consAddr)
	suite.True(found)

	infoResp, err = queryClient.SigningInfo(gocontext.Background(),
		&types.QuerySigningInfoRequest{ConsAddress: consAddr.String()})
	suite.NoError(err)
	suite.Equal(info, infoResp.ValSigningInfo)
}

func (suite *KeeperTestSuite) TestGRPCSigningInfos() {
	queryClient := suite.queryClient

	var signingInfos []types.ValidatorSigningInfo

	suite.slashingKeeper.IterateValidatorSigningInfos(suite.ctx, func(consAddr sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	})

	// verify all values are returned without pagination
	infoResp, err := queryClient.SigningInfos(gocontext.Background(),
		&types.QuerySigningInfosRequest{Pagination: nil})
	suite.NoError(err)
	suite.Equal(signingInfos, infoResp.Info)

	infoResp, err = queryClient.SigningInfos(gocontext.Background(),
		&types.QuerySigningInfosRequest{Pagination: &query.PageRequest{Limit: 1, CountTotal: true}})
	suite.NoError(err)
	suite.Len(infoResp.Info, 1)
	suite.Equal(signingInfos[0], infoResp.Info[0])
	suite.NotNil(infoResp.Pagination.NextKey)
	suite.Equal(uint64(2), infoResp.Pagination.Total)
}
