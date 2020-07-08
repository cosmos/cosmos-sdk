package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// define constants used for testing
const (
	testClientIDA = "testclientIDA"
	testClientIDB = "testClientIDb"

	testConnection = "testconnectionatob"
	testPort1      = "bank"
	testPort2      = "testportid"
	testChannel1   = "firstchannel"
	testChannel2   = "secondchannel"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

// define variables used for testing
var (
	testAddr1, _ = sdk.AccAddressFromBech32("cosmos1scqhwpgsmr6vmztaa7suurfl52my6nd2kmrudl")
	testAddr2, _ = sdk.AccAddressFromBech32("cosmos1scqhwpgsmr6vmztaa7suurfl52my6nd2kmrujl")

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
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	// reset prefixCoins at each setup
	prefixCoins = sdk.NewCoins(sdk.NewCoin("bank/firstchannel/atom", sdk.NewInt(100)))
	prefixCoins2 = sdk.NewCoins(sdk.NewCoin("testportid/secondchannel/atom", sdk.NewInt(100)))
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
