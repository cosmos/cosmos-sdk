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
	Route() string
	Type() string
	ValidateBasic() sdk.Error
	String() string
	Timeout() uint64
	MarshalAmino() (string, error)
	MarshalJSON() ([]byte, error)
}

type Channel struct {
	Port             string
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
