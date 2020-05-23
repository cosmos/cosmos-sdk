package ibc_test

import (
	lite "github.com/tendermint/tendermint/lite2"

	"github.com/cosmos/cosmos-sdk/x/ibc"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func (suite *IBCTestSuite) TestValidateGenesis() {
	counterparty, err := connection.NewCounterparty(clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix")))
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		genState ibc.GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: ibc.DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: ibc.GenesisState{
				ClientGenesis: client.NewGenesisState(
					[]exported.ClientState{
						ibctmtypes.NewClientState(clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
						localhosttypes.NewClientState("chaindID", 10),
					},
					[]client.ConsensusStates{
						client.NewClientConsensusStates(
							clientID,
							[]exported.ConsensusState{
								ibctmtypes.NewConsensusState(
									suite.header.Time, commitmenttypes.NewMerkleRoot(suite.header.AppHash), suite.header.GetHeight(), suite.header.ValidatorSet,
								),
							},
						),
					},
					true,
				),
				ConnectionGenesis: connection.NewGenesisState(
					[]connection.End{
						connection.NewConnectionEnd(connection.INIT, connectionID, clientID, counterparty, []string{"1.0.0"}),
					},
					[]connection.Paths{
						connection.NewConnectionPaths(clientID, []string{host.ConnectionPath(connectionID)}),
					},
				),
				ChannelGenesis: channel.NewGenesisState(
					[]channel.IdentifiedChannel{
						channel.NewIdentifiedChannel(
							port1, channel1, channel.NewChannel(
								channel.INIT, channelOrder,
								channel.NewCounterparty(port2, channel2), []string{connectionID}, channelVersion,
							),
						),
					},
					[]channel.PacketAckCommitment{
						channel.NewPacketAckCommitment(port2, channel2, 1, []byte("ack")),
					},
					[]channel.PacketAckCommitment{
						channel.NewPacketAckCommitment(port1, channel1, 1, []byte("commit_hash")),
					},
					[]channel.PacketSequence{
						channel.NewPacketSequence(port1, channel1, 1),
					},
					[]channel.PacketSequence{
						channel.NewPacketSequence(port2, channel2, 1),
					},
					[]channel.PacketSequence{
						channel.NewPacketSequence(port2, channel2, 1),
					},
				),
			},
			expPass: true,
		},
		{
			name: "invalid client genesis",
			genState: ibc.GenesisState{
				ClientGenesis: client.NewGenesisState(
					[]exported.ClientState{
						ibctmtypes.NewClientState(clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
						localhosttypes.NewClientState("(chaindID)", 0),
					},
					nil,
					false,
				),
				ConnectionGenesis: connection.DefaultGenesisState(),
			},
			expPass: false,
		},
		{
			name: "invalid connection genesis",
			genState: ibc.GenesisState{
				ClientGenesis: client.DefaultGenesisState(),
				ConnectionGenesis: connection.NewGenesisState(
					[]connection.End{
						connection.NewConnectionEnd(connection.INIT, connectionID, "(CLIENTIDONE)", counterparty, []string{"1.0.0"}),
					},
					[]connection.Paths{
						connection.NewConnectionPaths(clientID, []string{host.ConnectionPath(connectionID)}),
					},
				),
			},
			expPass: false,
		},
		{
			name: "invalid channel genesis",
			genState: ibc.GenesisState{
				ClientGenesis:     client.DefaultGenesisState(),
				ConnectionGenesis: connection.DefaultGenesisState(),
				ChannelGenesis: channel.GenesisState{
					Acknowledgements: []channel.PacketAckCommitment{
						channel.NewPacketAckCommitment("(portID)", channel1, 1, []byte("ack")),
					},
				},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
