package ibc_test

import (
	lite "github.com/tendermint/tendermint/lite2"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	"github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *IBCTestSuite) TestValidateGenesis() {
	testCases := []struct {
		name     string
		genState types.GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: types.DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: types.GenesisState{
				ClientGenesis: clienttypes.NewGenesisState(
					[]clienttypes.GenesisClientState{
						clienttypes.NewGenesisClientState(
							clientID, ibctmtypes.NewClientState(lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs()),
						),
						clienttypes.NewGenesisClientState(
							exported.ClientTypeLocalHost, localhosttypes.NewClientState("chaindID", 10),
						),
					},
					[]clienttypes.ClientConsensusStates{
						clienttypes.NewClientConsensusStates(
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
				ConnectionGenesis: connectiontypes.NewGenesisState(
					[]connectiontypes.IdentifiedConnection{
						connectiontypes.NewIdentifiedConnection(connectionID, connectiontypes.NewConnectionEnd(connectiontypes.INIT, clientID, connectiontypes.NewCounterparty(clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))), []string{ibctesting.ConnectionVersion})),
					},
					[]connectiontypes.ConnectionPaths{
						connectiontypes.NewConnectionPaths(clientID, []string{host.ConnectionPath(connectionID)}),
					},
				),
				ChannelGenesis: channeltypes.NewGenesisState(
					[]channeltypes.IdentifiedChannel{
						channeltypes.NewIdentifiedChannel(
							port1, channel1, channeltypes.NewChannel(
								channeltypes.INIT, channelOrder,
								channeltypes.NewCounterparty(port2, channel2), []string{connectionID}, channelVersion,
							),
						),
					},
					[]channeltypes.PacketAckCommitment{
						channeltypes.NewPacketAckCommitment(port2, channel2, 1, []byte("ack")),
					},
					[]channeltypes.PacketAckCommitment{
						channeltypes.NewPacketAckCommitment(port1, channel1, 1, []byte("commit_hash")),
					},
					[]channeltypes.PacketSequence{
						channeltypes.NewPacketSequence(port1, channel1, 1),
					},
					[]channeltypes.PacketSequence{
						channeltypes.NewPacketSequence(port2, channel2, 1),
					},
					[]channeltypes.PacketSequence{
						channeltypes.NewPacketSequence(port2, channel2, 1),
					},
				),
			},
			expPass: true,
		},
		{
			name: "invalid client genesis",
			genState: types.GenesisState{
				ClientGenesis: clienttypes.NewGenesisState(
					[]clienttypes.GenesisClientState{
						clienttypes.NewGenesisClientState(
							clientID, ibctmtypes.NewClientState(lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs()),
						),
						clienttypes.NewGenesisClientState(
							exported.ClientTypeLocalHost, localhosttypes.NewClientState("(chaindID)", 0),
						),
					},
					nil,
					false,
				),
				ConnectionGenesis: connectiontypes.DefaultGenesisState(),
			},
			expPass: false,
		},
		{
			name: "invalid connection genesis",
			genState: types.GenesisState{
				ClientGenesis: clienttypes.DefaultGenesisState(),
				ConnectionGenesis: connectiontypes.NewGenesisState(
					[]connectiontypes.IdentifiedConnection{
						connectiontypes.NewIdentifiedConnection(connectionID, connectiontypes.NewConnectionEnd(connectiontypes.INIT, "(CLIENTIDONE)", connectiontypes.NewCounterparty(clientID, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))), []string{"1.0.0"})),
					},
					[]connectiontypes.ConnectionPaths{
						connectiontypes.NewConnectionPaths(clientID, []string{host.ConnectionPath(connectionID)}),
					},
				),
			},
			expPass: false,
		},
		{
			name: "invalid channel genesis",
			genState: types.GenesisState{
				ClientGenesis:     clienttypes.DefaultGenesisState(),
				ConnectionGenesis: connectiontypes.DefaultGenesisState(),
				ChannelGenesis: channeltypes.GenesisState{
					Acknowledgements: []channeltypes.PacketAckCommitment{
						channeltypes.NewPacketAckCommitment("(portID)", channel1, 1, []byte("ack")),
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
