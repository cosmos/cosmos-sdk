package token

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

type PacketSend struct {
	ToAddress sdk.AccAddress `json:"to_address"`
	Amount    sdk.Coins      `json:"amount"`
}

var _ channel.Packet = PacketSend{}

func (packet PacketSend) Timeout() uint64 {
	return 1000000000 // TODO
}

func (packet PacketSend) Route() string {
	return "token"
}
