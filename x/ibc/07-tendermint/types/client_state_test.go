package types_test

import (
	ics23 "github.com/confio/ics23/go"

	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	testClientID     = "clientidone"
	testConnectionID = "connectionid"
	testPortID       = "testportid"
	testChannelID    = "testchannelid"
	testSequence     = 1
)

func (suite *TendermintTestSuite) TestValidate() {
	testCases := []struct {
		name        string
		clientState types.ClientState
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
		clientState    ibctmtypes.ClientState
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
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			nil, suite.cdc, suite.aminoCdc, tc.consensusState.Root, height, "chainA", tc.consensusState.GetHeight(), tc.prefix, tc.proof, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyConnectionState() {
	counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, commitmenttypes.NewMerklePrefix([]byte("ibc")))
	conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []string{"1.0.0"})

	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		connection     connectiontypes.ConnectionEnd
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          []byte
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, chainID, chainID, height, commitmenttypes.GetSDKSpecs()),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			connection:  conn,
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyConnectionState(
			nil, suite.cdc, height, tc.prefix, tc.proof, testConnectionID, tc.connection, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyChannelState() {
	counterparty := channeltypes.NewCounterparty(testPortID, testChannelID)
	ch := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")

	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		channel        channeltypes.Channel
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          []byte
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, height, commitmenttypes.GetSDKSpecs()),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			channel:     ch,
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyChannelState(
			nil, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, &tc.channel, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketCommitment() {
	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		commitment     []byte
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          []byte
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, height, commitmenttypes.GetSDKSpecs()),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			commitment:  []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketCommitment(
			nil, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.commitment, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgement() {
	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		ack            []byte
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          []byte
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			ack:         []byte{},
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgement(
			nil, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.ack, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          []byte
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, chainID, height, commitmenttypes.GetSDKSpecs()),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyPacketAcknowledgementAbsence(
			nil, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyNextSeqRecv() {
	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		consensusState ibctmtypes.ConsensusState
		prefix         commitmenttypes.MerklePrefix
		proof          []byte
		expPass        bool
	}{
		// FIXME: uncomment
		// {
		// 	name:         "successful verification",
		// 	clientState:  ibctmtypes.NewClientState(chainID, height, commitmenttypes.GetSDKSpecs()),
		// 	connection:   conn,
		// 	consensusState: ibctmtypes.ConsensusState{
		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
		// 	},
		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.MerklePrefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: ibctmtypes.ClientState{LatestHeight: height, FrozenHeight: height - 1},
			consensusState: ibctmtypes.ConsensusState{
				Root: commitmenttypes.NewMerkleRoot(suite.header.AppHash),
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: ibctmtypes.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			consensusState: ibctmtypes.ConsensusState{
				Root:         commitmenttypes.NewMerkleRoot(suite.header.AppHash),
				ValidatorSet: suite.valSet,
			},
			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
			proof:   []byte{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyNextSequenceRecv(
			nil, suite.cdc, height, tc.prefix, tc.proof, testPortID, testChannelID, testSequence, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
