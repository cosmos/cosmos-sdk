package types_test

import (
	ics23 "github.com/confio/ics23/go"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

const (
	testClientID     = "clientidone"
	testConnectionID = "connectionid"
	testPortID       = "testportid"
	testChannelID    = "testchannelid"
	testSequence     = 1
)

var (
	invalidProof = []byte("invalid proof")
)

func (suite *TendermintTestSuite) TestValidate() {
	testCases := []struct {
		name        string
		clientState *types.ClientState
		expPass     bool
	}{
		{
			name:        "valid client",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			expPass:     true,
		},
		{
			name:        "invalid chainID",
			clientState: ibctmtypes.NewClientState("  ", types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			expPass:     false,
		},
		{
			name:        "invalid trust level",
			clientState: ibctmtypes.NewClientState(chainID, types.Fraction{Numerator: 0, Denominator: 1}, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			expPass:     false,
		},
		{
			name:        "invalid trusting period",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, 0, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			expPass:     false,
		},
		{
			name:        "invalid unbonding period",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, 0, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			expPass:     false,
		},
		{
			name:        "invalid max clock drift",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, 0, height, commitmenttypes.GetSDKSpecs()),
			expPass:     false,
		},
		{
			name:        "invalid height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
			expPass:     false,
		},
		{
			name:        "trusting period not less than unbonding period",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, ubdPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			expPass:     false,
		},
		{
			name:        "proof specs is nil",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, ubdPeriod, ubdPeriod, maxClockDrift, height, nil),
			expPass:     false,
		},
		{
			name:        "proof specs contains nil",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, ubdPeriod, ubdPeriod, maxClockDrift, height, []*ics23.ProofSpec{ics23.TendermintSpec, nil}),
			expPass:     false,
		},
	}

	for _, tc := range testCases {
		err := tc.clientState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyClientConsensusState() {
	testCases := []struct {
		name           string
		clientState    *ibctmtypes.ClientState
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          []byte
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:        "successful verification",
		// 	clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: &ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root:               commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
				NextValidatorsHash: suite.valsHash,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			nil, suite.cdc, tc.consensusState.Root, height, "chainA", tc.consensusState.GetHeight(), tc.prefix, tc.proof, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

// test verification of the connection on chainB being represented in the
// light client on chainA
func (suite *TendermintTestSuite) TestVerifyConnectionState() {
	var (
		clientState *types.ClientState
		proof       []byte
		proofHeight uint64
		prefix      commitmenttypes.MerklePrefix
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"successful verification", func() {}, true,
		},
		{
			"ApplyPrefix failed", func() {
				prefix = commitmenttypes.MerklePrefix{}
			}, false,
		},
		{
			"latest client height < height", func() {
				proofHeight = clientState.LatestHeight + 1
			}, false,
		},
		{
			"client is frozen", func() {
				clientState.FrozenHeight = 1
			}, false,
		},
		{
			"proof verification failed", func() {
				proof = invalidProof
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			// setup testing conditions
			clientA, _, _, connB, _, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
			connection := suite.chainB.GetConnection(connB)

			var ok bool
			clientStateI := suite.chainA.GetClientState(clientA)
			clientState, ok = clientStateI.(*types.ClientState)
			suite.Require().True(ok)

			prefix = suite.chainB.GetPrefix()

			// make connection proof
			connectionKey := host.KeyConnection(connB.ID)
			proof, proofHeight = suite.chainB.QueryProof(connectionKey)

			tc.malleate() // make changes as necessary

			store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

			err := clientState.VerifyConnectionState(
				store, suite.chainA.Codec, proofHeight, &prefix, proof, connB.ID, connection,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// test verification of the channel on chainB being represented in the light
// client on chainA
func (suite *TendermintTestSuite) TestVerifyChannelState() {
	var (
		clientState *types.ClientState
		proof       []byte
		proofHeight uint64
		prefix      commitmenttypes.MerklePrefix
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"successful verification", func() {}, true,
		},
		{
			"ApplyPrefix failed", func() {
				prefix = commitmenttypes.MerklePrefix{}
			}, false,
		},
		{
			"latest client height < height", func() {
				proofHeight = clientState.LatestHeight + 1
			}, false,
		},
		{
			"client is frozen", func() {
				clientState.FrozenHeight = 1
			}, false,
		},
		{
			"proof verification failed", func() {
				proof = invalidProof
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			// setup testing conditions
			clientA, _, _, _, _, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			channel := suite.chainB.GetChannel(channelB)

			var ok bool
			clientStateI := suite.chainA.GetClientState(clientA)
			clientState, ok = clientStateI.(*types.ClientState)
			suite.Require().True(ok)

			prefix = suite.chainB.GetPrefix()

			// make channel proof
			channelKey := host.KeyChannel(channelB.PortID, channelB.ID)
			proof, proofHeight = suite.chainB.QueryProof(channelKey)

			tc.malleate() // make changes as necessary

			store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

			err := clientState.VerifyChannelState(
				store, suite.chainA.Codec, proofHeight, &prefix, proof,
				channelB.PortID, channelB.ID, channel,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// test verification of the packet commitment on chainB being represented
// in the light client on chainA. A send from chainB to chainA is simulated.
func (suite *TendermintTestSuite) TestVerifyPacketCommitment() {
	var (
		clientState *types.ClientState
		proof       []byte
		proofHeight uint64
		prefix      commitmenttypes.MerklePrefix
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"successful verification", func() {}, true,
		},
		{
			"ApplyPrefix failed", func() {
				prefix = commitmenttypes.MerklePrefix{}
			}, false,
		},
		{
			"latest client height < height", func() {
				proofHeight = clientState.LatestHeight + 1
			}, false,
		},
		{
			"client is frozen", func() {
				clientState.FrozenHeight = 1
			}, false,
		},
		{
			"proof verification failed", func() {
				proof = invalidProof
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			// setup testing conditions
			clientA, _, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, 100, 0)
			err := suite.coordinator.SendPacket(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			var ok bool
			clientStateI := suite.chainA.GetClientState(clientA)
			clientState, ok = clientStateI.(*types.ClientState)
			suite.Require().True(ok)

			prefix = suite.chainB.GetPrefix()

			// make packet commitment proof
			packetKey := host.KeyPacketCommitment(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			proof, proofHeight = suite.chainB.QueryProof(packetKey)

			tc.malleate() // make changes as necessary

			store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

			err = clientState.VerifyPacketCommitment(
				store, suite.chainA.Codec, proofHeight, &prefix, proof,
				packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(), channeltypes.CommitPacket(packet),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// test verification of the acknowledgement on chainB being represented
// in the light client on chainA. A send and ack from chainA to chainB
// is simulated.
func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgement() {
	var (
		clientState *types.ClientState
		proof       []byte
		proofHeight uint64
		prefix      commitmenttypes.MerklePrefix
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"successful verification", func() {}, true,
		},
		{
			"ApplyPrefix failed", func() {
				prefix = commitmenttypes.MerklePrefix{}
			}, false,
		},
		{
			"latest client height < height", func() {
				proofHeight = clientState.LatestHeight + 1
			}, false,
		},
		{
			"client is frozen", func() {
				clientState.FrozenHeight = 1
			}, false,
		},
		{
			"proof verification failed", func() {
				proof = invalidProof
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			// setup testing conditions
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100, 0)

			// send packet
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// write ack
			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			var ok bool
			clientStateI := suite.chainA.GetClientState(clientA)
			clientState, ok = clientStateI.(*types.ClientState)
			suite.Require().True(ok)

			prefix = suite.chainB.GetPrefix()

			// make packet acknowledgement proof
			acknowledgementKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight = suite.chainB.QueryProof(acknowledgementKey)

			tc.malleate() // make changes as necessary

			store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

			err = clientState.VerifyPacketAcknowledgement(
				store, suite.chainA.Codec, proofHeight, &prefix, proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(), ibctesting.TestHash,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// test verification of the absent acknowledgement on chainB being represented
// in the light client on chainA. A send from chainB to chainA is simulated, but
// no receive.
func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	var (
		clientState *types.ClientState
		proof       []byte
		proofHeight uint64
		prefix      commitmenttypes.MerklePrefix
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"successful verification", func() {}, true,
		},
		{
			"ApplyPrefix failed", func() {
				prefix = commitmenttypes.MerklePrefix{}
			}, false,
		},
		{
			"latest client height < height", func() {
				proofHeight = clientState.LatestHeight + 1
			}, false,
		},
		{
			"client is frozen", func() {
				clientState.FrozenHeight = 1
			}, false,
		},
		{
			"proof verification failed", func() {
				proof = invalidProof
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			// setup testing conditions
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100, 0)

			// send packet, but no recv
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// need to update chainA's client representing chainB to prove missing ack
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)

			var ok bool
			clientStateI := suite.chainA.GetClientState(clientA)
			clientState, ok = clientStateI.(*types.ClientState)
			suite.Require().True(ok)

			prefix = suite.chainB.GetPrefix()

			// make packet acknowledgement absence proof
			acknowledgementKey := host.KeyPacketAcknowledgement(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			proof, proofHeight = suite.chainB.QueryProof(acknowledgementKey)

			tc.malleate() // make changes as necessary

			store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

			err = clientState.VerifyPacketAcknowledgementAbsence(
				store, suite.chainA.Codec, proofHeight, &prefix, proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// test verification of the next receive sequence on chainB being represented
// in the light client on chainA. A send and receive from chainB to chainA is
// simulated.
func (suite *TendermintTestSuite) TestVerifyNextSeqRecv() {
	var (
		clientState *types.ClientState
		proof       []byte
		proofHeight uint64
		prefix      commitmenttypes.MerklePrefix
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"successful verification", func() {}, true,
		},
		{
			"ApplyPrefix failed", func() {
				prefix = commitmenttypes.MerklePrefix{}
			}, false,
		},
		{
			"latest client height < height", func() {
				proofHeight = clientState.LatestHeight + 1
			}, false,
		},
		{
			"client is frozen", func() {
				clientState.FrozenHeight = 1
			}, false,
		},
		{
			"proof verification failed", func() {
				proof = invalidProof
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			// setup testing conditions
			clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
			packet := channeltypes.NewPacket(ibctesting.TestHash, 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 100, 0)

			// send packet
			err := suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
			suite.Require().NoError(err)

			// write ack, next seq recv incremented
			err = suite.coordinator.PacketExecuted(suite.chainB, suite.chainA, packet, clientA)
			suite.Require().NoError(err)

			// need to update chainA's client representing chainB
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, clientexported.Tendermint)

			var ok bool
			clientStateI := suite.chainA.GetClientState(clientA)
			clientState, ok = clientStateI.(*types.ClientState)
			suite.Require().True(ok)

			prefix = suite.chainB.GetPrefix()

			// make next seq recv proof
			nextSeqRecvKey := host.KeyNextSequenceRecv(packet.GetDestPort(), packet.GetDestChannel())
			proof, proofHeight = suite.chainB.QueryProof(nextSeqRecvKey)

			tc.malleate() // make changes as necessary

			store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

			err = clientState.VerifyNextSequenceRecv(
				store, suite.chainA.Codec, proofHeight, &prefix, proof,
				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
