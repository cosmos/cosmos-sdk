package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// TestConnOpenInit - Chain A (ID #1) initializes (INIT state) a connection with
// Chain B (ID #2) which is yet UNINITIALIZED
func (suite *KeeperTestSuite) TestConnOpenInit() {
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
		}, true},
		{"connection already exists", func() {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
		}, false},
		{"couldn't add connection to client", func() {}, false},
	}

	counterparty := types.NewCounterparty(testClientIDB, testConnectionIDB, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenInit(suite.chainA.GetContext(), testConnectionIDA, testClientIDB, counterparty)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

// TestConnOpenTry - Chain B (ID #2) calls ConnOpenTry to verify the state of
// connection on Chain A (ID #1) is INIT
func (suite *KeeperTestSuite) TestConnOpenTry() {
	// counterparty for A on B
	counterparty := types.NewCounterparty(
		testClientIDB, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.chainB.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()),
	)

	testCases := []struct {
		msg      string
		malleate func() uint64
		expPass  bool
	}{
		{"success", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight() - 1
		}, true},
		{"consensus height > latest height", func() uint64 {
			return 0
		}, false},
		{"self consensus state not found", func() uint64 {
			//suite.ctx = suite.ctx.WithBlockHeight(100)
			return 100
		}, false},
		{"connection state verification invalid", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return 0
		}, false},
		{"consensus state verification invalid", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"invalid previous connection", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return 0
		}, false},
		{"couldn't add connection to client", func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return 0
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			consensusHeight := tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDA)
			proofInit, proofHeight := queryProof(suite.chainA, connectionKey)

			consensusKey := prefixedClientKey(testClientIDB, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := queryProof(suite.chainA, consensusKey)

			err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenTry(
				suite.chainB.GetContext(), testConnectionIDB, counterparty, testClientIDA,
				types.GetCompatibleVersions(), proofInit, proofConsensus,
				proofHeight+1, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed with consensus height %d and proof height %d: %s", i, consensusHeight, proofHeight, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed with consensus height %d and proof height %d: %s", i, consensusHeight, proofHeight, tc.msg)
			}
		})
	}
}

// TestConnOpenAck - Chain A (ID #1) calls TestConnOpenAck to acknowledge (ACK state)
// the initialization (TRYINIT) of the connection on  Chain B (ID #2).
func (suite *KeeperTestSuite) TestConnOpenAck() {
	version := types.GetCompatibleVersions()[0]

	testCases := []struct {
		msg      string
		version  string
		malleate func() uint64
		expPass  bool
	}{
		{"success", version, func() uint64 {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight()
		}, true},
		{"success from tryopen", version, func() uint64 {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.updateClient(suite.chainB)
			return suite.chainB.Header.GetHeight()
		}, true},
		{"consensus height > latest height", version, func() uint64 {
			return 10
		}, false},
		{"connection not found", version, func() uint64 {
			return 2
		}, false},
		{"connection state is not INIT", version, func() uint64 {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"incompatible IBC versions", "2.0", func() uint64 {
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"self consensus state not found", version, func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"connection state verification failed", version, func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
		{"consensus state verification failed", version, func() uint64 {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainB.updateClient(suite.chainA)
			return suite.chainB.Header.GetHeight()
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			consensusHeight := tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDB)
			proofTry, proofHeight := queryProof(suite.chainB, connectionKey)

			consensusKey := prefixedClientKey(testClientIDA, host.KeyConsensusState(consensusHeight))
			proofConsensus, _ := queryProof(suite.chainB, consensusKey)

			err := suite.chainA.App.IBCKeeper.ConnectionKeeper.ConnOpenAck(
				suite.chainA.GetContext(), testConnectionIDA, tc.version, proofTry, proofConsensus,
				proofHeight+1, consensusHeight,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed with consensus height %d: %s", i, consensusHeight, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed with consensus height %d: %s", i, consensusHeight, tc.msg)
			}
		})
	}
}

// TestConnOpenConfirm - Chain B (ID #2) calls ConnOpenConfirm to confirm that
// Chain A (ID #1) state is now OPEN.
func (suite *KeeperTestSuite) TestConnOpenConfirm() {
	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA, types.OPEN)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.TRYOPEN)
			suite.chainB.updateClient(suite.chainA)
		}, true},
		{"connection not found", func() {}, false},
		{"chain B's connection state is not TRYOPEN", func() {
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB, types.UNINITIALIZED)
			suite.chainA.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.OPEN)
			suite.chainA.updateClient(suite.chainB)
		}, false},
		{"connection state verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			suite.chainA.createConnection(testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, types.INIT)
			suite.chainB.createConnection(testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA, types.TRYOPEN)
			suite.chainA.updateClient(suite.chainA)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			connectionKey := host.KeyConnection(testConnectionIDA)
			proofAck, proofHeight := queryProof(suite.chainA, connectionKey)

			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, proofAck, proofHeight+1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ConnectionKeeper.ConnOpenConfirm(
					suite.chainB.GetContext(), testConnectionIDB, proofAck, proofHeight+1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
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
