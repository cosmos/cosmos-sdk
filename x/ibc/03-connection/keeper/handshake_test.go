package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// TestConnOpenInit - chainA initializes (INIT state) a connection with
// chainB which is yet UNINITIALIZED
func (suite *KeeperTestSuite) TestConnOpenInit() {
	var (
		clientA string
		clientB string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, true},
		{"connection already exists", func() {
			clientA, clientB, _, _ = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
		{"couldn't add connection to client", func() {
			// swap client identifiers to result in client that does not exist
			clientB, clientA = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)
			counterparty := types.NewCounterparty(clientB, connB.ID, suite.chainB.GetPrefix())

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.chainA.GetContext(), connA.ID, clientA, counterparty)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestConnOpenTry - chainB calls ConnOpenTry to verify the state of
// connection on chainA is INIT
func (suite *KeeperTestSuite) TestConnOpenTry() {
	var (
		clientA         string
		clientB         string
		consensusHeight uint64
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)
		}, true},
		{"consensus height > latest height", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			consensusHeight = uint64(suite.chainB.GetContext().BlockHeight()) + 1
		}, false},
		{"self consensus state not found", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			consensusHeight = 1
		}, false},
		{"connection state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			// chainA connection not created
		}, false},
		{"consensus state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			// give chainA wrong consensus state for chainB
			consState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainA.GetContext(), clientA)
			suite.Require().True(found)

			tmConsState, ok := consState.(ibctmtypes.ConsensusState)
			suite.Require().True(ok)

			tmConsState.Timestamp = time.Now()
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, tmConsState.Height, tmConsState)

			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)
		}, false},
		{"invalid previous connection is in TRYOPEN", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)

			// open init chainA
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// open try chainB
			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest()   // reset
			consensusHeight = 0 // must be explicity changed in malleate

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)
			counterparty := types.NewCounterparty(clientA, connA.ID, suite.chainA.GetPrefix())

			connectionKey := host.KeyConnection(connA.ID)
			proofInit, proofHeight := suite.chainA.QueryProof(connectionKey)

			// retrieve consensus state to provide proof for
			consState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainA.GetContext(), clientA)
			suite.Require().True(found)

			if consensusHeight == 0 {
				consensusHeight = consState.GetHeight()
			}
			consensusKey := host.FullKeyClientPath(clientA, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := suite.chainA.QueryProof(consensusKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.chainB.GetContext(), connB.ID, counterparty, clientB,
				types.GetCompatibleVersions(), proofInit, proofConsensus,
				proofHeight, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestConnOpenAck - Chain A (ID #1) calls TestConnOpenAck to acknowledge (ACK state)
// the initialization (TRYINIT) of the connection on  Chain B (ID #2).
func (suite *KeeperTestSuite) TestConnOpenAck() {
	var (
		clientA         string
		clientB         string
		consensusHeight uint64
		version         string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, true},
		{"success from tryopen", func() {
			// chainA is in TRYOPEN, chainB is in TRYOPEN
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connB, connA, err := suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)

			// set chainB to TRYOPEN
			connection := suite.chainB.GetConnection(connB)
			connection.State = types.TRYOPEN
			suite.chainB.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainB.GetContext(), connB.ID, connection)
			// update clientB so state change is committed
			suite.coordinator.UpdateClient(suite.chainB, suite.chainA, clientB, clientexported.Tendermint)

			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)
		}, true},
		{"consensus height > latest height", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			consensusHeight = uint64(suite.chainA.GetContext().BlockHeight()) + 1
		}, false},
		{"connection not found", func() {
			// connections are never created
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
		{"connection state is not INIT", func() {
			// connection state is already OPEN on chainA
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenAck(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)
		}, false},
		{"incompatible IBC versions", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			// set version to a non-compatible version
			version = "2.0"
		}, false},
		{"self consensus state not found", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			consensusHeight = 1
		}, false},
		{"connection state verification failed", func() {
			// chainB connection is not in INIT
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			_, _, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)
		}, false},
		{"consensus state verification failed", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// give chainB wrong consensus state for chainA
			consState, found := suite.chainB.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainB.GetContext(), clientB)
			suite.Require().True(found)

			tmConsState, ok := consState.(ibctmtypes.ConsensusState)
			suite.Require().True(ok)

			tmConsState.Timestamp = time.Now()
			suite.chainB.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainB.GetContext(), clientB, tmConsState.Height, tmConsState)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.msg, func() {
			suite.SetupTest()                          // reset
			version = types.GetCompatibleVersions()[0] // must be explicitly changed in malleate
			consensusHeight = 0                        // must be explicitly changed in malleate

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainB.GetFirstTestConnection(clientB, clientA)

			connectionKey := host.KeyConnection(connB.ID)
			proofTry, proofHeight := suite.chainB.QueryProof(connectionKey)

			// retrieve consensus state to provide proof for
			consState, found := suite.chainB.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainB.GetContext(), clientB)
			suite.Require().True(found)

			if consensusHeight == 0 {
				consensusHeight = consState.GetHeight()
			}
			consensusKey := host.FullKeyClientPath(clientB, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := suite.chainB.QueryProof(consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
				suite.chainA.GetContext(), connA.ID, version,
				proofTry, proofConsensus, proofHeight, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestConnOpenConfirm - chainB calls ConnOpenConfirm to confirm that
// chainA state is now OPEN.
func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	var (
		clientA string
		clientB string
	)
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenAck(suite.chainA, suite.chainB, connA, connB)
			suite.Require().NoError(err)
		}, true},
		{"connection not found", func() {
			// connections are never created
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			// connections are OPEN
			clientA, clientB, _, _ = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
		{"connection state verification failed", func() {
			// chainA is in INIT
			clientA, clientB = suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			connA, connB, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			err = suite.coordinator.ConnOpenTry(suite.chainB, suite.chainA, connB, connA)
			suite.Require().NoError(err)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			tc.malleate()

			connA := suite.chainA.GetFirstTestConnection(clientA, clientB)
			connB := suite.chainA.GetFirstTestConnection(clientB, clientA)

			connectionKey := host.KeyConnection(connA.ID)
			proofAck, proofHeight := suite.chainA.QueryProof(connectionKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
				suite.chainB.GetContext(), connB.ID, proofAck, proofHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestOpenTryVersionNegotiation uses tests the result of the version negotiation done during
// the handshake procedure OpenTry.
// TODO: following the refactor of this file based on ibc testing package, these cases
// should become individual cases for the OpenTry handshake.
// https://github.com/cosmos/cosmos-sdk/issues/5558
func TestOpenTryVersionNegotiation(t *testing.T) {
	testCases := []struct {
		name                 string
		counterpartyVersions []string
		expPass              bool
	}{
		{"valid counterparty versions", types.GetCompatibleVersions(), true},
		{"empty counterparty versions", []string{}, false},
		{"no counterparty match", []string{"(version won't match,[])"}, false},
	}

	// Test OpenTry variety of cases using counterpartyVersions
	for i, tc := range testCases {
		coordinator := ibctesting.NewCoordinator(t, 2)
		chainA := coordinator.GetChain(ibctesting.GetChainID(0))
		chainB := coordinator.GetChain(ibctesting.GetChainID(1))

		clientA, clientB := coordinator.SetupClients(chainA, chainB, clientexported.Tendermint)
		connA, connB, err := coordinator.ConnOpenInit(chainA, chainB, clientA, clientB)
		require.NoError(t, err)

		counterparty := types.NewCounterparty(clientA, connA.ID, chainA.GetPrefix())

		connectionKey := host.KeyConnection(connA.ID)
		proofInit, proofHeight := chainA.QueryProof(connectionKey)

		// retrieve consensus state to provide proof for
		consState, found := chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(chainA.GetContext(), clientA)
		require.True(t, found)

		consensusHeight := consState.GetHeight()
		consensusKey := host.FullKeyClientPath(clientA, host.KeyConsensusState(consensusHeight))
		proofConsensus, _ := chainA.QueryProof(consensusKey)

		err = chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(chainB.GetContext(),
			connB.ID, counterparty, clientB, tc.counterpartyVersions,
			proofInit, proofConsensus, proofHeight, consensusHeight,
		)

		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

// TestOpenAckVersionNegotiation uses tests the result of the version negotiation done during
// the handshake procedure OpenAck.
// TODO: following the refactor of this file based on ibc testing package, these cases
// should become individual cases for the OpenAck handshake.
// https://github.com/cosmos/cosmos-sdk/issues/5558
func TestOpenAckVersionNegotiation(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		expPass bool
	}{
		{"valid selected version", types.DefaultIBCVersion, true},
		{"empty selected versions", "", false},
		{"selected version not supported", "(not supported,[])", false},
	}

	// Test OpenTry variety of cases using counterpartyVersions
	for i, tc := range testCases {
		coordinator := ibctesting.NewCoordinator(t, 2)
		chainA := coordinator.GetChain(ibctesting.GetChainID(0))
		chainB := coordinator.GetChain(ibctesting.GetChainID(1))

		clientA, clientB := coordinator.SetupClients(chainA, chainB, clientexported.Tendermint)
		connA, connB, err := coordinator.ConnOpenInit(chainA, chainB, clientA, clientB)
		require.NoError(t, err)

		err = coordinator.ConnOpenTry(chainB, chainA, connB, connA)
		require.NoError(t, err)

		connectionKey := host.KeyConnection(connB.ID)
		proofTry, proofHeight := chainB.QueryProof(connectionKey)

		// retrieve consensus state to provide proof for
		consState, found := chainB.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(chainB.GetContext(), clientB)
		require.True(t, found)

		consensusHeight := consState.GetHeight()
		consensusKey := host.FullKeyClientPath(clientB, host.KeyConsensusState(consensusHeight))
		proofConsensus, _ := chainB.QueryProof(consensusKey)

		err = chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
			chainA.GetContext(), connA.ID, tc.version,
			proofTry, proofConsensus, proofHeight, consensusHeight,
		)

		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

type testCase = struct {
	msg      string
	malleate func()
	expPass  bool
}
