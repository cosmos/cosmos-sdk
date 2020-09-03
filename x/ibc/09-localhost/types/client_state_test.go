package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
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
			clientState: types.NewClientState("chainID", clienttypes.Height{}),
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
				suite.store.Set(host.KeyClientState(), bz)
			},
			counterparty: clientState,
			expPass:      true,
		},
		{
			name:        "proof verification failed: invalid client",
			clientState: clientState,
			malleate: func() {
				bz := clienttypes.MustMarshalClientState(suite.cdc, clientState)
				suite.store.Set(host.KeyClientState(), bz)
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
				suite.store, suite.cdc, nil, 10, nil, "", []byte{}, tc.counterparty,
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
		nil, nil, nil, 0, "", 0, nil, nil, nil,
	)
	suite.Require().NoError(err)
}
func (suite *LocalhostTestSuite) TestCheckHeaderAndUpdateState() {
	clientState := types.NewClientState("chainID", clientHeight)
	cs, _, err := clientState.CheckHeaderAndUpdateState(suite.ctx, nil, nil, nil)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.ctx.BlockHeight(), int64(cs.GetLatestHeight()))
	suite.Require().Equal(suite.ctx.BlockHeader().ChainID, clientState.ChainId)
}

func (suite *LocalhostTestSuite) TestVerifyConnectionState() {
	counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, commitmenttypes.NewMerklePrefix([]byte("ibc")))
	conn1 := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []string{"1.0.0"})
	conn2 := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []string{"2.0.0"})

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
				suite.store.Set(host.KeyConnection(testConnectionID), bz)
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
				suite.store.Set(host.KeyConnection(testConnectionID), []byte("connection"))
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
				suite.store.Set(host.KeyConnection(testConnectionID), bz)
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
				suite.store, suite.cdc, height, nil, []byte{}, testConnectionID, &tc.connection,
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
				suite.store.Set(host.KeyChannel(testPortID, testChannelID), bz)
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
				suite.store.Set(host.KeyChannel(testPortID, testChannelID), []byte("channel"))

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
				suite.store.Set(host.KeyChannel(testPortID, testChannelID), bz)

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
				suite.store, suite.cdc, height, nil, []byte{}, testPortID, testChannelID, &tc.channel,
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
					host.KeyPacketCommitment(testPortID, testChannelID, testSequence), []byte("commitment"),
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
					host.KeyPacketCommitment(testPortID, testChannelID, testSequence), []byte("different"),
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
				suite.store, suite.cdc, height, nil, []byte{}, testPortID, testChannelID, testSequence, tc.commitment,
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
					host.KeyPacketAcknowledgement(testPortID, testChannelID, testSequence), []byte("acknowledgement"),
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
					host.KeyPacketAcknowledgement(testPortID, testChannelID, testSequence), []byte("different"),
				)
			},
			ack:     []byte("acknowledgment"),
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
				suite.store, suite.cdc, height, nil, []byte{}, testPortID, testChannelID, testSequence, tc.ack,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *LocalhostTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	clientState := types.NewClientState("chainID", clientHeight)

	err := clientState.VerifyPacketAcknowledgementAbsence(
		suite.store, suite.cdc, height, nil, nil, testPortID, testChannelID, testSequence,
	)

	suite.Require().NoError(err, "ack absence failed")

	suite.store.Set(host.KeyPacketAcknowledgement(testPortID, testChannelID, testSequence), []byte("ack"))

	err = clientState.VerifyPacketAcknowledgementAbsence(
		suite.store, suite.cdc, height, nil, nil, testPortID, testChannelID, testSequence,
	)
	suite.Require().Error(err, "ack exists in store")
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
					host.KeyNextSequenceRecv(testPortID, testChannelID),
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
					host.KeyNextSequenceRecv(testPortID, testChannelID),
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
				suite.store, suite.cdc, height, nil, []byte{}, testPortID, testChannelID, nextSeqRecv,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
