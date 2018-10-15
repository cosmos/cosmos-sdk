package sentinel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgRegisterVpnService:
			return handleRegisterVpnService(ctx, k, msg)
		case MsgDeleteVpnUser:
			return handleDeleteVpnUser(ctx, k, msg)
		case MsgRegisterMasterNode:
			return handleMsgRegisterMasterNode(ctx, k, msg)
		case MsgDeleteMasterNode:
			return handleMsgDeleteMasterNode(ctx, k, msg)
		case MsgPayVpnService:
			return handleMsgPayVpnService(ctx, k, msg)
		case MsgGetVpnPayment:
			return handleMsgGetVpnPayment(ctx, k, msg)
		case MsgRefund:
			return handleMsgRefund(ctx, k, msg)
		case MsgSendTokens:
			return handleMsgMsgSendTokens(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("unrecognized message").Result()
		}
	}
}

func handleMsgRegisterMasterNode(ctx sdk.Context, keeper Keeper, msg MsgRegisterMasterNode) sdk.Result {
	_, err := keeper.RegisterMasterNode(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)
	return sdk.Result{
		Tags: msg.Tags(),
		Data: d,
	}
}

func handleMsgMsgSendTokens(ctx sdk.Context, keeper Keeper, msg MsgSendTokens) sdk.Result {
	address, err := keeper.SendTokens(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)
	tags := sdk.NewTags("Transfer to Address:", []byte(address.String()))
	return sdk.Result{
		Data: d,
		Tags: tags,
	}
}

func handleRegisterVpnService(ctx sdk.Context, keeper Keeper, msg MsgRegisterVpnService) sdk.Result {

	_, err := keeper.RegisterVpnService(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)

	tag := sdk.NewTags("vpn registered address", []byte(msg.From.String()))
	return sdk.Result{
		Data: d,
		Tags: tag,
	}
}

func handleDeleteVpnUser(ctx sdk.Context, keeper Keeper, msg MsgDeleteVpnUser) sdk.Result {
	_, err := keeper.DeleteVpnService(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)
	tag := sdk.NewTags("deleted Vpn address", []byte(msg.Vaddr))
	return sdk.Result{
		Data: d,
		Tags: tag,
	}
}

func handleMsgDeleteMasterNode(ctx sdk.Context, keeper Keeper, msg MsgDeleteMasterNode) sdk.Result {
	_, err := keeper.DeleteMasterNode(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)
	tag := sdk.NewTags("deleted MasterNode address", []byte(msg.Maddr))
	return sdk.Result{
		Data: d,
		Tags: tag,
	}
}

func handleMsgPayVpnService(ctx sdk.Context, keeper Keeper, msg MsgPayVpnService) sdk.Result {
	id, err := keeper.PayVpnService(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)
	tag := sdk.NewTags("sender address", []byte(msg.From.String())).AppendTag("seesion id", []byte(id))
	return sdk.Result{
		Data: d,
		Tags: tag,
	}
}

func handleMsgGetVpnPayment(ctx sdk.Context, keeper Keeper, msg MsgGetVpnPayment) sdk.Result {

	sessionid, clientAddr, err := keeper.GetVpnPayment(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)
	tags := sdk.NewTags("Vpn Provider Address:", []byte(msg.From.String())).AppendTag("seesionId", sessionid).AppendTag("Client Address", []byte(clientAddr.String()))
	return sdk.Result{
		Data: d,
		Tags: tags,
	}
}

func handleMsgRefund(ctx sdk.Context, keeper Keeper, msg MsgRefund) sdk.Result {
	address, err := keeper.RefundBal(ctx, msg)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalJSON(msg)
	tags := sdk.NewTags("client Refund Address:", []byte(address.String()))
	return sdk.Result{
		Data: d,
		Tags: tags,
	}
}
