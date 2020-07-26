package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	cdc := app.AppCodec()

	connectionID := "connectionidone"

	connection := types.ConnectionEnd{
		ClientID: "clientidone",
		Versions: []string{"1.0"},
	}

	paths := types.ClientPaths{
		Paths: []string{connectionID},
	}

	kvPairs := tmkv.Pairs{
		tmkv.Pair{
			Key:   host.KeyClientConnections(connection.ClientID),
			Value: cdc.MustMarshalBinaryBare(&paths),
		},
		tmkv.Pair{
			Key:   host.KeyConnection(connectionID),
			Value: cdc.MustMarshalBinaryBare(&connection),
		},
		tmkv.Pair{
			Key:   []byte{0x99},
			Value: []byte{0x99},
		},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ClientPaths", fmt.Sprintf("ClientPaths A: %v\nClientPaths B: %v", paths, paths)},
		{"ConnectionEnd", fmt.Sprintf("ConnectionEnd A: %v\nConnectionEnd B: %v", connection, connection)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			res, found := simulation.NewDecodeStore(cdc, kvPairs[i], kvPairs[i])
			if i == len(tests)-1 {
				require.False(t, found, string(kvPairs[i].Key))
				require.Empty(t, res, string(kvPairs[i].Key))
			} else {
				require.True(t, found, string(kvPairs[i].Key))
				require.Equal(t, tt.expectedLog, res, string(kvPairs[i].Key))
			}
		})
	}
}
