package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// define variables used for testing
var (
	testCoins, _ = sdk.ParseCoins("100atom")
	prefixCoins  = sdk.NewCoins(sdk.NewCoin("bank/firstchannel/atom", sdk.NewInt(100)))
	prefixCoins2 = sdk.NewCoins(sdk.NewCoin("testportid/secondchannel/atom", sdk.NewInt(100)))
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	sender   sdk.AccAddress
	receiver sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	suite.sender = sdk.AccAddress(tmtypes.NewMockPV().PrivKey.PubKey().Address())
	suite.receiver = sdk.AccAddress(tmtypes.NewMockPV().PrivKey.PubKey().Address())
}

func (suite *KeeperTestSuite) TestGetTransferAccount() {
	expectedMaccAddr := sdk.AccAddress(crypto.AddressHash([]byte(types.ModuleName)))

	macc := suite.chainA.App.TransferKeeper.GetTransferAccount(suite.chainA.GetContext())

	suite.Require().NotNil(macc)
	suite.Require().Equal(types.ModuleName, macc.GetName())
	suite.Require().Equal(expectedMaccAddr, macc.GetAddress())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
