package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type SoloMachineTestSuite struct {
	suite.Suite

	solomachine *ibctesting.Solomachine
	coordinator *ibctesting.Coordinator

	// testing chain used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	store sdk.KVStore
}

func (suite *SoloMachineTestSuite) SetupTest() {
	suite.solomachine = ibctesting.NewSolomachine(suite.T(), "testingsolomachine")
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	suite.store = suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientexported.ClientTypeSoloMachine)

	bz, err := codec.MarshalAny(suite.chainA.Codec, suite.solomachine.ClientState())
	suite.Require().NoError(err)
	suite.store.Set(host.KeyClientState(), bz)
}

func TestSoloMachineTestSuite(t *testing.T) {
	suite.Run(t, new(SoloMachineTestSuite))
}

func (suite *SoloMachineTestSuite) GetSequenceFromStore() uint64 {
	bz := suite.store.Get(host.KeyClientState())
	suite.Require().NotNil(bz)

	var clientState clientexported.ClientState
	err := codec.UnmarshalAny(suite.chainA.Codec, &clientState, bz)
	suite.Require().NoError(err)
	return clientState.GetLatestHeight()
}

func (suite *SoloMachineTestSuite) GetInvalidProof() []byte {
	invalidProof, err := suite.chainA.Codec.MarshalBinaryBare(&types.TimestampedSignature{Timestamp: suite.solomachine.Time})
	suite.Require().NoError(err)

	return invalidProof
}
