package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case ibc.MsgPacket:
			switch packet := msg.Packet.(type) {
			case SequencePacket:
				return handleMyPacket(ctx, k, packet, msg.ChannelID)
			}
		}
		return sdk.ErrUnknownRequest("21345").Result()
	}
}

func handleMyPacket(ctx sdk.Context, k Keeper, packet SequencePacket, chanid string) (res sdk.Result) {
	err := k.CheckAndSetSequence(ctx, chanid, packet.Sequence)
	if err != nil {
		res.Log = "Invalid sequence" // should not return error, set only log
	}
	return
}
