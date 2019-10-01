package mockrecv

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case ibc.MsgPacket:
			switch packet := msg.Packet.(type) {
			case types.PacketSequence:
				return handleMyPacket(ctx, k, packet, msg.ChannelID)
			default:
				return sdk.ErrUnknownRequest("23331345").Result()
			}
		default:
			return sdk.ErrUnknownRequest("21345").Result()
		}
	}
}

func handleMyPacket(ctx sdk.Context, k Keeper, packet types.PacketSequence, chanid string) (res sdk.Result) {
	err := k.UpdateSequence(ctx, chanid, packet.Sequence)
	if err != nil {
		res.Log = "Invalid sequence" // should not return error, set only log
	}
	return
}
