package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC Handler should not return error. Returning error will make the channel
// not able to be proceed. Send receipt packet when the packet is non executable.
type Handler func(ctx sdk.Context, packet Packet)

/*
type ReturnHandler func(ctx sdk.Context, packet sdk.Packet) sdk.Packet
*/

func WrapHandler(h Handler, sdkh sdk.Handler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgPacket:
			h(ctx, msg.Packet)
			return sdk.Result{Events:ctx.EventManager().Events()}
		default:
			return sdkh(ctx, msg)
		}
	}
}

// TODO
/*
func WrapReturnHandler(h ReturnHandler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		receipt := h(ctx, msg.Packet)	
		if receipt != nil {	
		// Send receipt packet to the receipt channel
		}
		
		return sdk.Result{Events: ctx.EventManager().Events()}
	}
}
*/
