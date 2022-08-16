package keeper_test

import (
	gocontext "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/slashing/testslashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *KeeperTestSuite) TestGRPCQueryParams() {
	queryClient := s.queryClient
	require := s.Require()

	paramsResp, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})

	require.NoError(err)
	require.Equal(testslashing.TestParams(), paramsResp.Params)
}

func (s *KeeperTestSuite) TestGRPCSigningInfo() {
	queryClient, ctx, keeper := s.queryClient, s.ctx, s.slashingKeeper
	require := s.Require()

	infoResp, err := queryClient.SigningInfo(gocontext.Background(), &types.QuerySigningInfoRequest{ConsAddress: ""})
	require.Error(err)
	require.Nil(infoResp)

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		0,
		int64(0),
		time.Unix(2, 0),
		false,
		int64(0),
	)

	keeper.SetValidatorSigningInfo(ctx, consAddr, signingInfo)
	info, found := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.True(found)

	infoResp, err = queryClient.SigningInfo(gocontext.Background(),
		&types.QuerySigningInfoRequest{ConsAddress: consAddr.String()})
	require.NoError(err)
	require.Equal(info, infoResp.ValSigningInfo)
}

func (s *KeeperTestSuite) TestGRPCSigningInfos() {
	queryClient, ctx, keeper := s.queryClient, s.ctx, s.slashingKeeper
	require := s.Require()

	// set two validator signing information
	consAddr1 := sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))
	consAddr2 := sdk.ConsAddress(sdk.AccAddress([]byte("addr2_______________")))
	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr1,
		0,
		int64(0),
		time.Unix(2, 0),
		false,
		int64(0),
	)

	keeper.SetValidatorSigningInfo(ctx, consAddr1, signingInfo)
	signingInfo.Address = string(consAddr2)
	keeper.SetValidatorSigningInfo(ctx, consAddr2, signingInfo)

	var signingInfos []types.ValidatorSigningInfo

	keeper.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	})

	// verify all values are returned without pagination
	infoResp, err := queryClient.SigningInfos(gocontext.Background(),
		&types.QuerySigningInfosRequest{Pagination: nil})
	require.NoError(err)
	require.Equal(signingInfos, infoResp.Info)

	infoResp, err = queryClient.SigningInfos(gocontext.Background(),
		&types.QuerySigningInfosRequest{Pagination: &query.PageRequest{Limit: 1, CountTotal: true}})
	require.NoError(err)
	require.Len(infoResp.Info, 1)
	require.Equal(signingInfos[0], infoResp.Info[0])
	require.NotNil(infoResp.Pagination.NextKey)
	require.Equal(uint64(2), infoResp.Pagination.Total)
}
