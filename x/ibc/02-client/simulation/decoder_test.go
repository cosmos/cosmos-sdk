package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/simulation"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	cdc := app.Codec()
	clientID := "clientidone"

	clientState := ibctmtypes.ClientState{
		FrozenHeight: 10,
	}

	consState := ibctmtypes.ConsensusState{
		Height:    10,
		Timestamp: time.Now().UTC(),
	}

	kvPairs := kv.Pairs{
		kv.Pair{
			Key:   host.FullKeyClientPath(clientID, host.KeyClientState()),
			Value: cdc.MustMarshalBinaryBare(clientState),
		},
		kv.Pair{
			Key:   host.FullKeyClientPath(clientID, host.KeyClientType()),
			Value: []byte(exported.Tendermint.String()),
		},
		kv.Pair{
			Key:   host.FullKeyClientPath(clientID, host.KeyConsensusState(10)),
			Value: cdc.MustMarshalBinaryBare(consState),
		},
		kv.Pair{
			Key:   []byte{0x99},
			Value: []byte{0x99},
		},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ClientState", fmt.Sprintf("ClientState A: %v\nClientState B: %v", clientState, clientState)},
		{"client type", fmt.Sprintf("Client type A: %s\nClient type B: %s", exported.Tendermint, exported.Tendermint)},
		{"ConsensusState", fmt.Sprintf("ConsensusState A: %v\nConsensusState B: %v", consState, consState)},
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
