package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	cdc := app.AppCodec()

	channelID := "channelidone"
	portID := "portidone"

	channel := types.Channel{
		State:   types.OPEN,
		Version: "1.0",
	}

	bz := []byte{0x1, 0x2, 0x3}

	kvPairs := tmkv.Pairs{
		tmkv.Pair{
			Key:   host.KeyChannel(portID, channelID),
			Value: cdc.MustMarshalBinaryBare(&channel),
		},
		tmkv.Pair{
			Key:   host.KeyNextSequenceSend(portID, channelID),
			Value: sdk.Uint64ToBigEndian(1),
		},
		tmkv.Pair{
			Key:   host.KeyNextSequenceRecv(portID, channelID),
			Value: sdk.Uint64ToBigEndian(1),
		},
		tmkv.Pair{
			Key:   host.KeyNextSequenceAck(portID, channelID),
			Value: sdk.Uint64ToBigEndian(1),
		},
		tmkv.Pair{
			Key:   host.KeyPacketCommitment(portID, channelID, 1),
			Value: bz,
		},
		tmkv.Pair{
			Key:   host.KeyPacketAcknowledgement(portID, channelID, 1),
			Value: bz,
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
		{"Channel", fmt.Sprintf("Channel A: %v\nChannel B: %v", channel, channel)},
		{"NextSeqSend", "NextSeqSend A: 1\nNextSeqSend B: 1"},
		{"NextSeqRecv", "NextSeqRecv A: 1\nNextSeqRecv B: 1"},
		{"NextSeqAck", "NextSeqAck A: 1\nNextSeqAck B: 1"},
		{"CommitmentHash", fmt.Sprintf("CommitmentHash A: %X\nCommitmentHash B: %X", bz, bz)},
		{"AckHash", fmt.Sprintf("AckHash A: %X\nAckHash B: %X", bz, bz)},
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
