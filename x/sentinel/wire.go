package sentinel

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgRegisterVpnService{}, "sentinel/registervpn", nil)
	//cdc.RegisterConcrete(MsgQueryRegisteredVpnService{}, "sentinel/queryvpnservice", nil)
	cdc.RegisterConcrete(MsgDeleteVpnUser{}, "sentienl/deletevpnservice", nil)
	cdc.RegisterConcrete(MsgRegisterMasterNode{}, "sentinel/masternoderegistration", nil)
	// cdc.RegisterConcrete(MsgQueryFromMasterNode{}, "sentienl/querythevpnservice", nil)
	cdc.RegisterConcrete(MsgDeleteMasterNode{}, "sentinel/deletemasternode", nil)
	cdc.RegisterConcrete(MsgPayVpnService{}, "sentinel/payvpnservice", nil)
	cdc.RegisterConcrete(MsgRefund{}, "sentinel/clientrefund", nil)
	cdc.RegisterConcrete(MsgGetVpnPayment{}, "sentinel/getvpnpayment", nil)
	cdc.RegisterConcrete(MsgSendTokens{}, "sentinel/sendtoken", nil)
}

var msgCdc = wire.NewCodec()

func init() {
	RegisterWire(msgCdc)
	wire.RegisterCrypto(msgCdc)
}
