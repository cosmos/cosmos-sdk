package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TransferPacketData struct {
	Amount   sdk.Coin
	Sender   sdk.AccAddress
	Receiver string
	Source   bool
}

func (packet TransferPacketData) MarshalAmino() ([]byte, error) {
	return MouduleCdc.MarshalBinaryBare(packet)
}

func (packet *TransferPacketData) UnmarshalAmino(bz []byte) (err error) {
	return MouduleCdc.UnmarshalBinaryBare(bz, packet)
}

func (packet TransferPacketData) Marshal() []byte {
	return MouduleCdc.MustMarshalBinaryBare(packet)
}

func (packet TransferPacketData) MarshalJSON() ([]byte, error) {
	return MouduleCdc.MarshalJSON(packet)
}

func (packet *TransferPacketData) UnmarshalJSON(bz []byte) (err error) {
	return MouduleCdc.UnmarshalJSON(bz, packet)
}

func (packet TransferPacketData) String() string {
	return fmt.Sprintf(`TransferPacketData:
	Amount:               %s
	Sender:               %s
	Receiver:             %s
	Source:               %v`,
		packet.Amount.String(),
		packet.Sender.String(),
		packet.Receiver,
		packet.Source,
	)
}
