package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

const (
	testPort1         = "firstport"
	testPort2         = "secondport"
	testConnectionIDA = "connectionidatob"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"

	testChannelOrder   = types.ORDERED
	testChannelVersion = "1.0"
)

func TestValidateGenesis(t *testing.T) {
	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
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
			genState: types.NewGenesisState(
				[]types.IdentifiedChannel{
					types.NewIdentifiedChannel(
						testPort1, testChannel1, types.NewChannel(
							types.INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
					types.NewIdentifiedChannel(
						testPort2, testChannel2, types.NewChannel(
							types.INIT, testChannelOrder, counterparty1, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
				[]types.PacketAckCommitment{
					types.NewPacketAckCommitment(testPort2, testChannel2, 1, []byte("ack")),
				},
				[]types.PacketAckCommitment{
					types.NewPacketAckCommitment(testPort1, testChannel1, 1, []byte("commit_hash")),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
				[]types.PacketSequence{
					types.NewPacketSequence(testPort2, testChannel2, 1),
				},
			),
			expPass: true,
		},
		{
			name: "invalid channel",
			genState: types.GenesisState{
				Channels: []types.IdentifiedChannel{
					types.NewIdentifiedChannel(
						testPort1, "(testChannel1)", types.NewChannel(
							types.INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
			},
			expPass: false,
		},
		{
			name: "invalid ack",
			genState: types.GenesisState{
				Acknowledgements: []types.PacketAckCommitment{
					types.NewPacketAckCommitment(testPort2, testChannel2, 1, nil),
				},
			},
			expPass: false,
		},
		{
			name: "invalid commitment",
			genState: types.GenesisState{
				Commitments: []types.PacketAckCommitment{
					types.NewPacketAckCommitment(testPort1, testChannel1, 1, nil),
				},
			},
			expPass: false,
		},
		{
			name: "invalid send seq",
			genState: types.GenesisState{
				SendSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, testChannel1, 0),
				},
			},
			expPass: false,
		},
		{
			name: "invalid recv seq",
			genState: types.GenesisState{
				RecvSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, "(testChannel1)", 1),
				},
			},
			expPass: false,
		},
		{
			name: "invalid recv seq 2",
			genState: types.GenesisState{
				RecvSequences: []types.PacketSequence{
					types.NewPacketSequence("(testPort1)", testChannel1, 1),
				},
			},
			expPass: false,
		},
		{
			name: "invalid ack seq",
			genState: types.GenesisState{
				AckSequences: []types.PacketSequence{
					types.NewPacketSequence(testPort1, "(testChannel1)", 1),
				},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
