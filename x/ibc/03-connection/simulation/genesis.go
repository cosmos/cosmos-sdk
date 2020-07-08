package simulation

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// GenConnectionGenesis returns the default connection genesis state.
func GenConnectionGenesis(_ *rand.Rand, _ []simtypes.Account) types.GenesisState {
	//return types.DefaultGenesisState()

	// Copied this in from a unit test
	var (
		connectionID  = "connectionidone"
		clientID      = "clientidone"
		connectionID2 = "connectionidtwo"
		clientID2     = "clientidtwo"
	)

	return types.NewGenesisState(
		[]types.ConnectionEnd{
			types.NewConnectionEnd(types.INIT, connectionID, clientID, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{types.DefaultIBCVersion}),
		},
		[]types.ConnectionPaths{
			{clientID, []string{host.ConnectionPath(connectionID)}},
		},
	)
}
