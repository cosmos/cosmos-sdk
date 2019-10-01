package types

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

var _ ibc.Packet = PacketSequence{}

type PacketSequence struct {
	Sequence uint64
}

func (packet PacketSequence) MarshalAmino() (string, error) {
	return fmt.Sprintf("sequence-packet-%d", packet.Sequence), nil
}

func (packet *PacketSequence) UnmarshalAmino(text string) (err error) {
	if !strings.HasPrefix(text, "sequence-packet-") {
		return errors.New("invalid PacketSequence string")
	}
	packet.Sequence, err = strconv.ParseUint(strings.TrimPrefix(text, "sequence-packet-"), 10, 64)
	return
}

func (packet PacketSequence) Marshal() []byte {
	cdc := codec.New()
	RegisterCodec(cdc)
	return cdc.MustMarshalBinaryBare(packet)
}

func (packet PacketSequence) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"sequence-packet-%d\"", packet.Sequence)), nil
}

func (packet *PacketSequence) UnmarshalJSON(bz []byte) (err error) {
	bz = bz[1 : len(bz)-1]
	if !bytes.HasPrefix(bz, []byte("sequence-packet-")) {
		return errors.New("invalid PacketSequence string")
	}
	packet.Sequence, err = strconv.ParseUint(strings.TrimPrefix(string(bz), "sequence-packet-"), 10, 64)
	return
}

func (PacketSequence) SenderPort() string {
	return "ibcmocksend"
}

func (PacketSequence) ReceiverPort() string {
	return "ibcmockrecv"
}

func (PacketSequence) String() string {
	return "sequence-packet"
}

func (PacketSequence) Timeout() uint64 {
	return 0
}

func (PacketSequence) Type() string {
	return "empty-packet"
}

func (PacketSequence) ValidateBasic() sdk.Error {
	return nil
}
