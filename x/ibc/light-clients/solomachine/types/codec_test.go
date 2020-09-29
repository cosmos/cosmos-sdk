package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite SoloMachineTestSuite) TestCanUnmarshalDataByType() {
	var (
		data []byte
		err  error
	)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		cdc := suite.chainA.App.AppCodec()
		cases := []struct {
			name     string
			dataType types.DataType
			malleate func()
			expPass  bool
		}{
			{
				"empty data", types.CLIENT, func() {
					data = []byte{}
				}, false,
			},
			{
				"unspecified", types.UNSPECIFIED, func() {
					path := solomachine.GetClientStatePath(counterpartyClientIdentifier)
					data, err = types.ClientStateDataBytes(cdc, path, solomachine.ClientState())
					suite.Require().NoError(err)
				}, false,
			},
			{
				"client", types.CLIENT, func() {
					path := solomachine.GetClientStatePath(counterpartyClientIdentifier)
					data, err = types.ClientStateDataBytes(cdc, path, solomachine.ClientState())
					suite.Require().NoError(err)
				}, true,
			},
			{
				"bad client (provides consensus state data)", types.CLIENT, func() {
					path := solomachine.GetConsensusStatePath(counterpartyClientIdentifier, clienttypes.NewHeight(0, 5))
					data, err = types.ConsensusStateDataBytes(cdc, path, solomachine.ConsensusState())
					suite.Require().NoError(err)
				}, false,
			},
			{
				"consensus", types.CONSENSUS, func() {
					path := solomachine.GetConsensusStatePath(counterpartyClientIdentifier, clienttypes.NewHeight(0, 5))
					data, err = types.ConsensusStateDataBytes(cdc, path, solomachine.ConsensusState())
					suite.Require().NoError(err)

				}, true,
			},
			{
				"bad consensus (provides client state data)", types.CONSENSUS, func() {
					path := solomachine.GetClientStatePath(counterpartyClientIdentifier)
					data, err = types.ClientStateDataBytes(cdc, path, solomachine.ClientState())
					suite.Require().NoError(err)
				}, false,
			},
			{
				"connection", types.CONNECTION, func() {
					counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, prefix)
					conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, []string{"1.0.0"})
					path := solomachine.GetConnectionStatePath("connectionID")

					data, err = types.ConnectionStateDataBytes(cdc, path, conn)
					suite.Require().NoError(err)

				}, true,
			},
			{
				"channel", types.CHANNEL, func() {
					counterparty := channeltypes.NewCounterparty(testPortID, testChannelID)
					ch := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")
					path := solomachine.GetChannelStatePath("portID", "channelID")

					data, err = types.ChannelStateDataBytes(cdc, path, ch)
					suite.Require().NoError(err)
				}, true,
			},
			{
				"packet commitment", types.PACKETCOMMITMENT, func() {
					commitment := []byte("packet commitment")
					path := solomachine.GetPacketCommitmentPath("portID", "channelID")

					data, err = types.PacketCommitmentDataBytes(cdc, path, commitment)
					suite.Require().NoError(err)
				}, true,
			},
			{
				"packet acknowledgement", types.PACKETACKNOWLEDGEMENT, func() {
					commitment := []byte("packet acknowledgement")
					path := solomachine.GetPacketAcknowledgementPath("portID", "channelID")

					data, err = types.PacketAcknowledgementDataBytes(cdc, path, commitment)
					suite.Require().NoError(err)
				}, true,
			},
			{
				"packet acknowledgement absence", types.PACKETACKNOWLEDGEMENTABSENCE, func() {
					commitment := []byte("packet acknowledgement")
					path := solomachine.GetPacketAcknowledgementPath("portID", "channelID")

					data, err = types.PacketAcknowledgementDataBytes(cdc, path, commitment)
					suite.Require().NoError(err)
				}, true,
			},
			{
				"next sequence recv", types.NEXTSEQUENCERECV, func() {
					path := solomachine.GetNextSequenceRecvPath("portID", "channelID")

					data, err = types.NextSequenceRecvDataBytes(cdc, path, 10)
					suite.Require().NoError(err)
				}, true,
			},
		}

		for _, tc := range cases {
			tc := tc

			suite.Run(tc.name, func() {
				tc.malleate()

				res := types.CanUnmarshalDataByType(cdc, tc.dataType, data)

				suite.Require().Equal(tc.expPass, res)
			})
		}
	}

}
