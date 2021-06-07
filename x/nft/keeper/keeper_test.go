package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
	addrs       []sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = simapp.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
	suite.addrs = simapp.AddTestAddrsIncremental(suite.app, suite.ctx, 3, sdk.NewInt(30000000))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.NFTkeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) TestGetNFT() {
	nft := types.NFT{
		Id:    "painting1",
		Owner: suite.addrs[0].String(),
		Data:  nil,
	}
	suite.app.NFTkeeper.SetNFT(suite.ctx, nft)

	expect, has := suite.app.NFTkeeper.GetNFT(suite.ctx, nft.Id)
	suite.Require().True(has)
	suite.Require().EqualValues(expect, nft)
}

func (suite *KeeperTestSuite) TestIteratorNFTsByOwner() {
	nfts := []types.NFT{
		{
			Id:    "painting1",
			Owner: suite.addrs[0].String(),
		},
		{
			Id:    "painting2",
			Owner: suite.addrs[0].String(),
		},
		{
			Id:    "painting3",
			Owner: suite.addrs[0].String(),
		},
	}

	for _, nft := range nfts {
		suite.app.NFTkeeper.SetNFT(suite.ctx, nft)
	}

	var expectdNFTs []types.NFT
	suite.app.NFTkeeper.IteratorNFTsByOwner(suite.ctx, suite.addrs[0], func(nft types.NFT) {
		expectdNFTs = append(expectdNFTs, nft)
	})
	suite.Require().EqualValues(expectdNFTs, nfts)
}

func (suite *KeeperTestSuite) TestTransferOwnership() {
	ctx := suite.ctx
	nft := types.NFT{
		Id:    "painting1",
		Owner: suite.addrs[0].String(),
		Data:  nil,
	}
	suite.app.NFTkeeper.SetNFT(ctx, nft)

	err := suite.app.NFTkeeper.TransferOwnership(ctx, nft.Id, suite.addrs[0], suite.addrs[1])
	suite.Require().NoError(err)

	var expectdNFTs []types.NFT
	suite.app.NFTkeeper.IteratorNFTsByOwner(suite.ctx, suite.addrs[0], func(nft types.NFT) {
		expectdNFTs = append(expectdNFTs, nft)
	})
	suite.Require().Len(expectdNFTs, 0)

	suite.app.NFTkeeper.IteratorNFTsByOwner(suite.ctx, suite.addrs[1], func(nft types.NFT) {
		expectdNFTs = append(expectdNFTs, nft)
	})
	suite.Require().Len(expectdNFTs, 1)
	suite.Require().Equal(expectdNFTs[0].Owner, suite.addrs[1].String())
}

func (suite *KeeperTestSuite) TestRemoveNFT() {
	nft := types.NFT{
		Id:    "painting1",
		Owner: suite.addrs[0].String(),
		Data:  nil,
	}
	suite.app.NFTkeeper.SetNFT(suite.ctx, nft)

	expect, has := suite.app.NFTkeeper.GetNFT(suite.ctx, nft.Id)
	suite.Require().True(has)
	suite.Require().EqualValues(expect, nft)

	err := suite.app.NFTkeeper.RemoveNFT(suite.ctx, nft.Id)
	suite.Require().NoError(err)

	_, has = suite.app.NFTkeeper.GetNFT(suite.ctx, nft.Id)
	suite.Require().False(has)
}
