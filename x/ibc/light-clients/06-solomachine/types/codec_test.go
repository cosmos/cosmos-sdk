package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite SoloMachineTestSuite) TestUnmarshalDataByType() {
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
					conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, connectiontypes.ExportedVersionsToProto(connectiontypes.GetCompatibleVersions()), 0)
					path := solomachine.GetConnectionStatePath("connectionID")

					data, err = types.ConnectionStateDataBytes(cdc, path, conn)
					suite.Require().NoError(err)

				}, true,
			},
			{
				"bad connection (uses channel data)", types.CONNECTION, func() {
					counterparty := channeltypes.NewCounterparty(testPortID, testChannelID)
					ch := channeltypes.NewChannel(channeltypes.OPEN, channeltypes.ORDERED, counterparty, []string{testConnectionID}, "1.0.0")
					path := solomachine.GetChannelStatePath("portID", "channelID")

					data, err = types.ChannelStateDataBytes(cdc, path, ch)
					suite.Require().NoError(err)
				}, false,
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
				"bad channel (uses connection data)", types.CHANNEL, func() {
					counterparty := connectiontypes.NewCounterparty("clientB", testConnectionID, prefix)
					conn := connectiontypes.NewConnectionEnd(connectiontypes.OPEN, "clientA", counterparty, connectiontypes.ExportedVersionsToProto(connectiontypes.GetCompatibleVersions()), 0)
					path := solomachine.GetConnectionStatePath("connectionID")

					data, err = types.ConnectionStateDataBytes(cdc, path, conn)
					suite.Require().NoError(err)

				}, false,
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
				"bad packet commitment (uses next seq recv)", types.PACKETCOMMITMENT, func() {
					path := solomachine.GetNextSequenceRecvPath("portID", "channelID")

					data, err = types.NextSequenceRecvDataBytes(cdc, path, 10)
					suite.Require().NoError(err)
				}, false,
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
				"bad packet acknowledgement (uses next sequence recv)", types.PACKETACKNOWLEDGEMENT, func() {
					path := solomachine.GetNextSequenceRecvPath("portID", "channelID")

					data, err = types.NextSequenceRecvDataBytes(cdc, path, 10)
					suite.Require().NoError(err)
				}, false,
			},
			{
				"packet acknowledgement absence", types.PACKETRECEIPTABSENCE, func() {
					path := solomachine.GetPacketReceiptPath("portID", "channelID")

					data, err = types.PacketReceiptAbsenceDataBytes(cdc, path)
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
			{
				"bad next sequence recv (uses packet commitment)", types.NEXTSEQUENCERECV, func() {
					commitment := []byte("packet commitment")
					path := solomachine.GetPacketCommitmentPath("portID", "channelID")

					data, err = types.PacketCommitmentDataBytes(cdc, path, commitment)
					suite.Require().NoError(err)
				}, false,
			},
		}

		for _, tc := range cases {
			tc := tc

			suite.Run(tc.name, func() {
				tc.malleate()

				data, err := types.UnmarshalDataByType(cdc, tc.dataType, data)

				if tc.expPass {
					suite.Require().NoError(err)
					suite.Require().NotNil(data)
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(data)
				}
			})
		}
	}

}
