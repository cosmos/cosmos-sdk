package ibc_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *IBCTestSuite) TestValidateGenesis() {
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
						ibctmtypes.NewClientState(clientID, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
						localhosttypes.NewClientState(suite.store, "chaindID", 10),
					},
					[]client.ClientConsensusStates{
						client.NewClientConsensusStates(
							clientID,
							[]exported.ConsensusState{
								ibctmtypes.NewConsensusState(
									suite.header.Time, commitmenttypes.NewMerkleRoot(suite.header.AppHash), suite.header.GetHeight(), suite.header.ValidatorSet,
								),
							},
						),
					},
				),
				ConnectionGenesis: connection.NewGenesisState(
					[]connection.ConnectionEnd{
						connection.NewConnectionEnd(connectionexported.INIT, connectionID, clientID, connection.NewCounterparty(clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))), []string{"1.0.0"}),
					},
					[]connection.ConnectionPaths{
						connection.NewConnectionPaths(clientID, []string{ibctypes.ConnectionPath(connectionID)}),
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
						ibctmtypes.NewClientState(clientID, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
						localhosttypes.NewClientState(suite.store, "chaindID", 0),
					},
					nil,
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
					[]connection.ConnectionEnd{
						connection.NewConnectionEnd(connectionexported.INIT, connectionID, "CLIENTIDONE", connection.NewCounterparty(clientID, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))), []string{"1.0.0"}),
					},
					[]connection.ConnectionPaths{
						connection.NewConnectionPaths(clientID, []string{ibctypes.ConnectionPath(connectionID)}),
					},
				),
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
