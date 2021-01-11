package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/09-localhost/types"
)

const (
	testConnectionID = "connectionid"
	testPortID       = "testportid"
	testChannelID    = "testchannelid"
	testSequence     = 1
)

func (suite *LocalhostTestSuite) TestValidate() {
	testCases := []struct {
		name        string
		clientState *types.ClientState
		expPass     bool
	}{
		{
			name:        "valid client",
			clientState: types.NewClientState("chainID", clienttypes.NewHeight(3, 10)),
			expPass:     true,
		},
		{
			name:        "invalid chain id",
			clientState: types.NewClientState(" ", clienttypes.NewHeight(3, 10)),
			expPass:     false,
		},
		{
			name:        "invalid height",
			clientState: types.NewClientState("chainID", clienttypes.ZeroHeight()),
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

func (suite *LocalhostTestSuite) TestInitialize() {
	testCases := []struct {
		name      string
		consState exported.ConsensusState
		expPass   bool
	}{
		{
			"valid initialization",
			nil,
			true,
		},
		{
			"invalid consenus state",
			&ibctmtypes.ConsensusState{},
			false,
		},
	}

	clientState := types.NewClientState("chainID", clienttypes.NewHeight(3, 10))

	for _, tc := range testCases {
		err := clientState.Initialize(suite.ctx, suite.cdc, suite.store, tc.consState)

		if tc.expPass {
			suite.Require().NoError(err, "valid testcase: %s failed", tc.name)
		} else {
			suite.Require().Error(err, "invalid testcase: %s passed", tc.name)
		}
	}
}

func (suite *LocalhostTestSuite) TestVerifyClientState() {
	clientState := types.NewClientState("chainID", clientHeight)
	invalidClient := types.NewClientState("chainID", clienttypes.NewHeight(0, 12))

	testCases := []struct {
		name         string
		clientState  *types.ClientState
		malleate     func()
		counterparty *types.ClientState
		expPass      bool
	}{
		{
			name:        "proof verification success",
			clientState: clientState,
			malleate: func() {
				bz := clienttypes.MustMarshalClientState(suite.cdc, clientState)
				suite.store.Set(host.ClientStateKey(), bz)
			},
			counterparty: clientState,
			expPass:      true,
		},
		{
			name:        "proof verification failed: invalid client",
			clientState: clientState,
			malleate: func() {
				bz := clienttypes.MustMarshalClientState(suite.cdc, clientState)
				suite.store.Set(host.ClientStateKey(), bz)
			},
			counterparty: invalidClient,
			expPass:      false,
		},
		{
			name:         "proof verification failed: client not stored",
			clientState:  clientState,
			malleate:     func() {},
			counterparty: clientState,
			expPass:      false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			err := tc.clientState.VerifyClientState(
				suite.store, suite.cdc, clienttypes.NewHeight(0, 10), nil, "", []byte{}, tc.counterparty,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

func (suite *LocalhostTestSuite) TestVerifyClientConsensusState() {
	clientState := types.NewClientState("chainID", clientHeight)
	err := clientState.VerifyClientConsensusState(
		nil, nil, nil, "", nil, nil, nil, nil,
	)
	suite.Require().NoError(err)
}

func (suite *LocalhostTestSuite) TestCheckHeaderAndUpdateState() {
	clientState := types.NewClientState("chainID", clientHeight)
	cs, _, err := clientState.CheckHeaderAndUpdateState(suite.ctx, nil, nil, nil)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), cs.GetLatestHeight().GetRevisionNumber())
	suite.Require().Equal(suite.ctx.BlockHeight(), int64(cs.GetLatestHeight().GetRevisionHeight()))
	suite.Require().Equal(suite.ctx.BlockHeader().ChainID, clientState.ChainId)
}

func (suite *LocalhostTestSuite) TestMisbehaviourAndUpdateState() {
	clientState := types.NewClientState("chainID", clientHeight)
	cs, err := clientState.CheckMisbehaviourAndUpdateState(suite.ctx, nil, nil, nil)
	suite.Require().Error(err)
	suite.Require().Nil(cs)
}

func (suite *LocalhostTestSuite) TestProposedHeaderAndUpdateState() {
	clientState := types.NewClientState("chainID", clientHeight)
	cs, consState, err := clientState.CheckProposedHeaderAndUpdateState(suite.ctx, nil, nil, nil)
	suite.Require().Error(err)
	suite.Require().Nil(cs)
	suite.Require().Nil(consState)
}

func (suite *LocalhostTestSuite) TestVerifyConnectionState() {
	counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, commitmenttypes.NewMerklePrefix([]byte("ibc")))
	conn1 := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []*connectiontypes.Version{connectiontypes.NewVersion("1", nil)}, 0)
	conn2 := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []*connectiontypes.Version{connectiontypes.NewVersion("2", nil)}, 0)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		malleate    func()
		connection  connectiontypes.ConnectionEnd
		expPass     bool
	}{
		{
			name:        "proof verification success",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				bz, err := suite.cdc.MarshalBinaryBare(&conn1)
				suite.Require().NoError(err)
				suite.store.Set(host.ConnectionKey(testConnectionID), bz)
			},
			connection: conn1,
			expPass:    true,
		},
		{
			name:        "proof verification failed: connection not stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate:    func() {},
			connection:  conn1,
			expPass:     false,
		},
		{
			name:        "proof verification failed: unmarshal error",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(host.ConnectionKey(testConnectionID), []byte("connection"))
			},
			connection: conn1,
			expPass:    false,
		},
		{
			name:        "proof verification failed: different connection stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				bz, err := suite.cdc.MarshalBinaryBare(&conn2)
				suite.Require().NoError(err)
				suite.store.Set(host.ConnectionKey(testConnectionID), bz)
			},
			connection: conn1,
			expPass:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			err := tc.clientState.VerifyConnectionState(
				suite.store, suite.cdc, clientHeight, nil, []byte{}, testConnectionID, &tc.connection,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LocalhostTestSuite) TestVerifyChannelState() {
	counterparty := channeltypes.NewCounterparty(testPortID, testChannelID)
	ch1 := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")
	ch2 := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "2.0.0")

	testCases := []struct {
		name        string
		clientState *types.ClientState
		malleate    func()
		channel     channeltypes.Channel
		expPass     bool
	}{
		{
			name:        "proof verification success",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				bz, err := suite.cdc.MarshalBinaryBare(&ch1)
				suite.Require().NoError(err)
				suite.store.Set(host.ChannelKey(testPortID, testChannelID), bz)
			},
			channel: ch1,
			expPass: true,
		},
		{
			name:        "proof verification failed: channel not stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate:    func() {},
			channel:     ch1,
			expPass:     false,
		},
		{
			name:        "proof verification failed: unmarshal failed",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(host.ChannelKey(testPortID, testChannelID), []byte("channel"))

			},
			channel: ch1,
			expPass: false,
		},
		{
			name:        "proof verification failed: different channel stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				bz, err := suite.cdc.MarshalBinaryBare(&ch2)
				suite.Require().NoError(err)
				suite.store.Set(host.ChannelKey(testPortID, testChannelID), bz)

			},
			channel: ch1,
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			err := tc.clientState.VerifyChannelState(
				suite.store, suite.cdc, clientHeight, nil, []byte{}, testPortID, testChannelID, &tc.channel,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LocalhostTestSuite) TestVerifyPacketCommitment() {
	testCases := []struct {
		name        string
		clientState *types.ClientState
		malleate    func()
		commitment  []byte
		expPass     bool
	}{
		{
			name:        "proof verification success",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(
					host.PacketCommitmentKey(testPortID, testChannelID, testSequence), []byte("commitment"),
				)
			},
			commitment: []byte("commitment"),
			expPass:    true,
		},
		{
			name:        "proof verification failed: different commitment stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(
					host.PacketCommitmentKey(testPortID, testChannelID, testSequence), []byte("different"),
				)
			},
			commitment: []byte("commitment"),
			expPass:    false,
		},
		{
			name:        "proof verification failed: no commitment stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate:    func() {},
			commitment:  []byte{},
			expPass:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			err := tc.clientState.VerifyPacketCommitment(
				suite.store, suite.cdc, clientHeight, 0, 0, nil, []byte{}, testPortID, testChannelID, testSequence, tc.commitment,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LocalhostTestSuite) TestVerifyPacketAcknowledgement() {
	testCases := []struct {
		name        string
		clientState *types.ClientState
		malleate    func()
		ack         []byte
		expPass     bool
	}{
		{
			name:        "proof verification success",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(
					host.PacketAcknowledgementKey(testPortID, testChannelID, testSequence), []byte("acknowledgement"),
				)
			},
			ack:     []byte("acknowledgement"),
			expPass: true,
		},
		{
			name:        "proof verification failed: different ack stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(
					host.PacketAcknowledgementKey(testPortID, testChannelID, testSequence), []byte("different"),
				)
			},
			ack:     []byte("acknowledgement"),
			expPass: false,
		},
		{
			name:        "proof verification failed: no commitment stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate:    func() {},
			ack:         []byte{},
			expPass:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			err := tc.clientState.VerifyPacketAcknowledgement(
				suite.store, suite.cdc, clientHeight, 0, 0, nil, []byte{}, testPortID, testChannelID, testSequence, tc.ack,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LocalhostTestSuite) TestVerifyPacketReceiptAbsence() {
	clientState := types.NewClientState("chainID", clientHeight)

	err := clientState.VerifyPacketReceiptAbsence(
		suite.store, suite.cdc, clientHeight, 0, 0, nil, nil, testPortID, testChannelID, testSequence,
	)

	suite.Require().NoError(err, "receipt absence failed")

	suite.store.Set(host.PacketReceiptKey(testPortID, testChannelID, testSequence), []byte("receipt"))

	err = clientState.VerifyPacketReceiptAbsence(
		suite.store, suite.cdc, clientHeight, 0, 0, nil, nil, testPortID, testChannelID, testSequence,
	)
	suite.Require().Error(err, "receipt exists in store")
}

func (suite *LocalhostTestSuite) TestVerifyNextSeqRecv() {
	nextSeqRecv := uint64(5)

	testCases := []struct {
		name        string
		clientState *types.ClientState
		malleate    func()
		nextSeqRecv uint64
		expPass     bool
	}{
		{
			name:        "proof verification success",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(
					host.NextSequenceRecvKey(testPortID, testChannelID),
					sdk.Uint64ToBigEndian(nextSeqRecv),
				)
			},
			nextSeqRecv: nextSeqRecv,
			expPass:     true,
		},
		{
			name:        "proof verification failed: different nextSeqRecv stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate: func() {
				suite.store.Set(
					host.NextSequenceRecvKey(testPortID, testChannelID),
					sdk.Uint64ToBigEndian(3),
				)
			},
			nextSeqRecv: nextSeqRecv,
			expPass:     false,
		},
		{
			name:        "proof verification failed: no nextSeqRecv stored",
			clientState: types.NewClientState("chainID", clientHeight),
			malleate:    func() {},
			nextSeqRecv: nextSeqRecv,
			expPass:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			err := tc.clientState.VerifyNextSequenceRecv(
				suite.store, suite.cdc, clientHeight, 0, 0, nil, []byte{}, testPortID, testChannelID, nextSeqRecv,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
