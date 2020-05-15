package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testPort1         = "firstport"
	testPort2         = "secondport"
	testConnectionIDA = "connectionidatob"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"

	testChannelOrder   = ORDERED
	testChannelVersion = "1.0"
)

func TestValidateGenesis(t *testing.T) {
	counterparty1 := NewCounterparty(testPort1, testChannel1)
	counterparty2 := NewCounterparty(testPort2, testChannel2)
	testCases := []struct {
		name     string
		genState GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: NewGenesisState(
				[]IdentifiedChannel{
					NewIdentifiedChannel(
						testPort1, testChannel1, NewChannel(
							INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
					NewIdentifiedChannel(
						testPort2, testChannel2, NewChannel(
							INIT, testChannelOrder, counterparty1, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
				[]PacketAckCommitment{
					NewPacketAckCommitment(testPort2, testChannel2, 1, []byte("ack")),
				},
				[]PacketAckCommitment{
					NewPacketAckCommitment(testPort1, testChannel1, 1, []byte("commit_hash")),
				},
				[]PacketSequence{
					NewPacketSequence(testPort1, testChannel1, 1),
				},
				[]PacketSequence{
					NewPacketSequence(testPort2, testChannel2, 1),
				},
			),
			expPass: true,
		},
		{
			name: "invalid channel",
			genState: GenesisState{
				Channels: []IdentifiedChannel{
					NewIdentifiedChannel(
						testPort1, "(testChannel1)", NewChannel(
							INIT, testChannelOrder, counterparty2, []string{testConnectionIDA}, testChannelVersion,
						),
					),
				},
			},
			expPass: false,
		},
		{
			name: "invalid ack",
			genState: GenesisState{
				Acknowledgements: []PacketAckCommitment{
					NewPacketAckCommitment(testPort2, testChannel2, 1, nil),
				},
			},
			expPass: false,
		},
		{
			name: "invalid commitment",
			genState: GenesisState{
				Commitments: []PacketAckCommitment{
					NewPacketAckCommitment(testPort1, testChannel1, 1, nil),
				},
			},
			expPass: false,
		},
		{
			name: "invalid send seq",
			genState: GenesisState{
				SendSequences: []PacketSequence{
					NewPacketSequence(testPort1, testChannel1, 0),
				},
			},
			expPass: false,
		},
		{
			name: "invalid recv seq",
			genState: GenesisState{
				RecvSequences: []PacketSequence{
					NewPacketSequence(testPort1, "(testChannel1)", 1),
				},
			},
			expPass: false,
		},
		{
			name: "invalid recv seq 2",
			genState: GenesisState{
				RecvSequences: []PacketSequence{
					NewPacketSequence("(testPort1)", testChannel1, 1),
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
