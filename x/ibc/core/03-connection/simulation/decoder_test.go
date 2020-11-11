package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	cdc := app.AppCodec()

	connectionID := "connectionidone"

	connection := types.ConnectionEnd{
		ClientId: "clientidone",
		Versions: types.ExportedVersionsToProto(types.GetCompatibleVersions()),
	}

	paths := types.ClientPaths{
		Paths: []string{connectionID},
	}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{
				Key:   host.ClientConnectionsKey(connection.ClientId),
				Value: cdc.MustMarshalBinaryBare(&paths),
			},
			{
				Key:   host.ConnectionKey(connectionID),
				Value: cdc.MustMarshalBinaryBare(&connection),
			},
			{
				Key:   []byte{0x99},
				Value: []byte{0x99},
			},
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
			res, found := simulation.NewDecodeStore(cdc, kvPairs.Pairs[i], kvPairs.Pairs[i])
			if i == len(tests)-1 {
				require.False(t, found, string(kvPairs.Pairs[i].Key))
				require.Empty(t, res, string(kvPairs.Pairs[i].Key))
			} else {
				require.True(t, found, string(kvPairs.Pairs[i].Key))
				require.Equal(t, tt.expectedLog, res, string(kvPairs.Pairs[i].Key))
			}
		})
	}
}
