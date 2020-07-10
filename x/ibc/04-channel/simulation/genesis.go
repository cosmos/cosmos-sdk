package simulation

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
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

// GenChannelGenesis returns the default channel genesis state.
func GenChannelGenesis(_ *rand.Rand, _ []simtypes.Account) types.GenesisState {
	//return types.DefaultGenesisState()

	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)

	return types.NewGenesisState(
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
	)
}
