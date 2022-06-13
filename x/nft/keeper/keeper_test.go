package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmtime "github.com/tendermint/tendermint/libs/time"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/testutil"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

const (
	testClassID          = "kitty"
	testClassName        = "Crypto Kitty"
	testClassSymbol      = "kitty"
	testClassDescription = "Crypto Kitty"
	testClassURI         = "class uri"
	testClassURIHash     = "ae702cefd6b6a65fe2f991ad6d9969ed"
	testID               = "kitty1"
	testURI              = "kitty uri"
	testURIHash          = "229bfd3c1b431c14a526497873897108"
)

type TestSuite struct {
	suite.Suite

	ctx         sdk.Context
	addrs       []sdk.AccAddress
	queryClient nft.QueryClient
	nftKeeper   keeper.Keeper
}

func (s *TestSuite) SetupTest() {
	var (
		interfaceRegistry codectypes.InterfaceRegistry
		bankKeeper        bankkeeper.Keeper
		stakingKeeper     stakingkeeper.Keeper
		nftKeeper         keeper.Keeper
	)

	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&interfaceRegistry,
		&nftKeeper,
		&bankKeeper,
		&stakingKeeper,
	)
	s.Require().NoError(err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	nft.RegisterQueryServer(queryHelper, s.nftKeeper)
	queryClient := nft.NewQueryClient(queryHelper)

	s.ctx = ctx
	s.queryClient = queryClient
	s.addrs = simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 3, sdk.NewInt(30000000))
	s.nftKeeper = nftKeeper
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestSaveClass() {
	except := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, except)
	s.Require().NoError(err)

	actual, has := s.nftKeeper.GetClass(s.ctx, testClassID)
	s.Require().True(has)
	s.Require().EqualValues(except, actual)

	classes := s.nftKeeper.GetClasses(s.ctx)
	s.Require().EqualValues([]*nft.Class{&except}, classes)
}

func (s *TestSuite) TestUpdateClass() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	noExistClass := nft.Class{
		Id:          "kitty1",
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}

	err = s.nftKeeper.UpdateClass(s.ctx, noExistClass)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "nft class does not exist")

	except := nft.Class{
		Id:          testClassID,
		Name:        "My crypto Kitty",
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}

	err = s.nftKeeper.UpdateClass(s.ctx, except)
	s.Require().NoError(err)

	actual, has := s.nftKeeper.GetClass(s.ctx, testClassID)
	s.Require().True(has)
	s.Require().EqualValues(except, actual)
}

func (s *TestSuite) TestMint() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.nftKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	// test GetNFT
	actNFT, has := s.nftKeeper.GetNFT(s.ctx, testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(expNFT, actNFT)

	// test GetOwner
	owner := s.nftKeeper.GetOwner(s.ctx, testClassID, testID)
	s.Require().True(s.addrs[0].Equals(owner))

	// test GetNFTsOfClass
	actNFTs := s.nftKeeper.GetNFTsOfClass(s.ctx, testClassID)
	s.Require().EqualValues([]nft.NFT{expNFT}, actNFTs)

	// test GetNFTsOfClassByOwner
	actNFTs = s.nftKeeper.GetNFTsOfClassByOwner(s.ctx, testClassID, s.addrs[0])
	s.Require().EqualValues([]nft.NFT{expNFT}, actNFTs)

	// test GetBalance
	balance := s.nftKeeper.GetBalance(s.ctx, testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(1), balance)

	// test GetTotalSupply
	supply := s.nftKeeper.GetTotalSupply(s.ctx, testClassID)
	s.Require().EqualValues(uint64(1), supply)

	expNFT2 := nft.NFT{
		ClassId: testClassID,
		Id:      testID + "2",
		Uri:     testURI + "2",
	}
	err = s.nftKeeper.Mint(s.ctx, expNFT2, s.addrs[0])
	s.Require().NoError(err)

	// test GetNFTsOfClassByOwner
	actNFTs = s.nftKeeper.GetNFTsOfClassByOwner(s.ctx, testClassID, s.addrs[0])
	s.Require().EqualValues([]nft.NFT{expNFT, expNFT2}, actNFTs)

	// test GetBalance
	balance = s.nftKeeper.GetBalance(s.ctx, testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(2), balance)
}

func (s *TestSuite) TestBurn() {
	except := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, except)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.nftKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	err = s.nftKeeper.Burn(s.ctx, testClassID, testID)
	s.Require().NoError(err)

	// test GetNFT
	_, has := s.nftKeeper.GetNFT(s.ctx, testClassID, testID)
	s.Require().False(has)

	// test GetOwner
	owner := s.nftKeeper.GetOwner(s.ctx, testClassID, testID)
	s.Require().Nil(owner)

	// test GetNFTsOfClass
	actNFTs := s.nftKeeper.GetNFTsOfClass(s.ctx, testClassID)
	s.Require().Empty(actNFTs)

	// test GetNFTsOfClassByOwner
	actNFTs = s.nftKeeper.GetNFTsOfClassByOwner(s.ctx, testClassID, s.addrs[0])
	s.Require().Empty(actNFTs)

	// test GetBalance
	balance := s.nftKeeper.GetBalance(s.ctx, testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(0), balance)

	// test GetTotalSupply
	supply := s.nftKeeper.GetTotalSupply(s.ctx, testClassID)
	s.Require().EqualValues(uint64(0), supply)
}

func (s *TestSuite) TestUpdate() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	myNFT := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.nftKeeper.Mint(s.ctx, myNFT, s.addrs[0])
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     "updated",
	}

	err = s.nftKeeper.Update(s.ctx, expNFT)
	s.Require().NoError(err)

	// test GetNFT
	actNFT, has := s.nftKeeper.GetNFT(s.ctx, testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(expNFT, actNFT)
}

func (s *TestSuite) TestTransfer() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.nftKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	// valid owner
	err = s.nftKeeper.Transfer(s.ctx, testClassID, testID, s.addrs[1])
	s.Require().NoError(err)

	// test GetOwner
	owner := s.nftKeeper.GetOwner(s.ctx, testClassID, testID)
	s.Require().Equal(s.addrs[1], owner)

	balanceAddr0 := s.nftKeeper.GetBalance(s.ctx, testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(0), balanceAddr0)

	balanceAddr1 := s.nftKeeper.GetBalance(s.ctx, testClassID, s.addrs[1])
	s.Require().EqualValues(uint64(1), balanceAddr1)

	// test GetNFTsOfClassByOwner
	actNFTs := s.nftKeeper.GetNFTsOfClassByOwner(s.ctx, testClassID, s.addrs[1])
	s.Require().EqualValues([]nft.NFT{expNFT}, actNFTs)
}

func (s *TestSuite) TestExportGenesis() {
	class := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.nftKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.nftKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	expGenesis := &nft.GenesisState{
		Classes: []*nft.Class{&class},
		Entries: []*nft.Entry{{
			Owner: s.addrs[0].String(),
			Nfts:  []*nft.NFT{&expNFT},
		}},
	}
	genesis := s.nftKeeper.ExportGenesis(s.ctx)
	s.Require().Equal(expGenesis, genesis)
}

func (s *TestSuite) TestInitGenesis() {
	expClass := nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	expNFT := nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	expGenesis := &nft.GenesisState{
		Classes: []*nft.Class{&expClass},
		Entries: []*nft.Entry{{
			Owner: s.addrs[0].String(),
			Nfts:  []*nft.NFT{&expNFT},
		}},
	}
	s.nftKeeper.InitGenesis(s.ctx, expGenesis)

	actual, has := s.nftKeeper.GetClass(s.ctx, testClassID)
	s.Require().True(has)
	s.Require().EqualValues(expClass, actual)

	// test GetNFT
	actNFT, has := s.nftKeeper.GetNFT(s.ctx, testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(expNFT, actNFT)
}
