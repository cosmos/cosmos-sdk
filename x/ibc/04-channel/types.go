package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
type Packet struct {
	Sequence      uint64
	TimeoutHeight uint64

	SourceConnection string
	SourceChannel    string
	DestConnection   string
	DestChannel      string

	Data []byte
}
*/

type Packet interface {
	SenderPort() string
	ReceiverPort() string // == Route()
	Type() string
	ValidateBasic() sdk.Error
	Timeout() uint64
	Marshal() []byte // Should exclude PortID/ChannelID info
}

type Channel struct {
	Counterparty     string
	CounterpartyPort string
	ConnectionHops   []string
}

func (ch Channel) CounterpartyHops() (res []string) {
	res = make([]string, len(ch.ConnectionHops))
	for i, hop := range ch.ConnectionHops {
		res[len(res)-(i+1)] = hop
	}
	return
}
