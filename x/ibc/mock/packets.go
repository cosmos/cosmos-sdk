package mock

import (
	"errors"
	"fmt"
	"strings"
	"strconv"
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

var _ ibc.Packet = SequencePacket{}

type SequencePacket struct {
	Sequence uint64
}

func (packet SequencePacket) MarshalAmino() (string, error) {
	return fmt.Sprintf("sequence-packet-%d", packet.Sequence), nil
}

func (packet *SequencePacket) UnmarshalAmino(text string) (err error) {
	if !strings.HasPrefix(text, "sequence-packet-") {
		return errors.New("invalid SequencePacket string")
	}
	packet.Sequence, err = strconv.ParseUint(strings.TrimPrefix(text, "sequence-packet-"), 10, 64)
	return
}

func (packet SequencePacket) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"sequence-packet-%d\"", packet.Sequence)), nil
}

func (packet *SequencePacket) UnmarshalJSON(bz []byte) (err error) {
	bz = bz[1:len(bz)-1]
	if !bytes.HasPrefix(bz, []byte("sequence-packet-")) {
		return errors.New("invalid SequencePacket string")
	}
	packet.Sequence, err = strconv.ParseUint(strings.TrimPrefix(string(bz), "sequence-packet-"), 10, 64)
	return
}

func (SequencePacket) Route() string {
	return "ibc-mock"
}

func (SequencePacket) String() string {
	return "sequence-packet"
}

func (SequencePacket) Timeout() uint64 {
	return 0
}

func (SequencePacket) Type() string {
	return "empty-packet"
}

func (SequencePacket) ValidateBasic() sdk.Error {
	return nil
}
