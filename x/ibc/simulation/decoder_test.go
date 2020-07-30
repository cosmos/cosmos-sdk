package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/simulation"
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	cdc := app.AppCodec()
	aminoCdc := app.Codec()

	dec := simulation.NewDecodeStore(app.IBCKeeper.Codecs())

	clientID := "clientidone"
	connectionID := "connectionidone"
	channelID := "channelidone"
	portID := "portidone"

	clientState := ibctmtypes.ClientState{
		FrozenHeight: 10,
	}
	connection := connectiontypes.ConnectionEnd{
		ClientID: "clientidone",
		Versions: []string{"1.0"},
	}
	channel := channeltypes.Channel{
		State:   channeltypes.OPEN,
		Version: "1.0",
	}

	kvPairs := kv.Pairs{
		kv.Pair{
			Key:   host.FullKeyClientPath(clientID, host.KeyClientState()),
			Value: aminoCdc.MustMarshalBinaryBare(clientState),
		},
		kv.Pair{
			Key:   host.KeyConnection(connectionID),
			Value: cdc.MustMarshalBinaryBare(&connection),
		},
		kv.Pair{
			Key:   host.KeyChannel(portID, channelID),
			Value: cdc.MustMarshalBinaryBare(&channel),
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
		{"ConnectionEnd", fmt.Sprintf("ConnectionEnd A: %v\nConnectionEnd B: %v", connection, connection)},
		{"Channel", fmt.Sprintf("Channel A: %v\nChannel B: %v", channel, channel)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			if i == len(tests)-1 {
				require.Panics(t, func() { dec(kvPairs[i], kvPairs[i]) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
