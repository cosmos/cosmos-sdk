package sentinel

import (
	"encoding/json"
	"net"
	"reflect"
	"strconv"
	"strings"

	//log "github.com/logger"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

//

//
//
//
//

/// USE gofmt command for styling/structing the go code

type MsgRegisterVpnService struct {
	From     sdk.AccAddress
	Ip       string
	Netspeed int64
	Ppgb     int64
	Location string
}

func NewMsgRegisterVpnService(address sdk.AccAddress, ip string, ppgb int64, netspeed int64, location string) MsgRegisterVpnService {
	return MsgRegisterVpnService{
		From:     address,
		Ip:       ip,
		Ppgb:     ppgb,
		Netspeed: netspeed,
		Location: location,
	}
}

func validateIp(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 || len(parts) > 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}

	}
	ip := net.ParseIP(host)
	if ip.IsLoopback() || ip.IsMulticast() {
		return false
	}
	return true
}

func (msc MsgRegisterVpnService) Type() string {
	return "sentinel"
}

func (msc MsgRegisterVpnService) GetSignBytes() []byte {
	var byteformat []byte
	byteformat, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byteformat
}
func (msc MsgRegisterVpnService) ValidateBasic() sdk.Error {
	var a int64
	var s string
	if msc.From == nil {
		return sdk.ErrInvalidAddress("Invalid Address")
	}
	if reflect.TypeOf(msc.Ppgb) != reflect.TypeOf(a) || msc.Ppgb <= 0 || msc.Ppgb > 1000 {

		return ErrInvalidPricePerGb("Price per GB is not Valid")
	}
	if msc.Ip == "" || !validateIp(msc.Ip) || reflect.TypeOf(msc.Ip) != reflect.TypeOf(s) {

		return ErrInvalidIpAdress("Invalid IP address")
	}
	if reflect.TypeOf(msc.Netspeed) != reflect.TypeOf(a) || msc.Netspeed <= 0 || msc.Netspeed >= 1000 {
		return ErrInvalidNetspeed("NetSpeed is not Valid")
	}
	if msc.Location == "" || reflect.TypeOf(msc.Location) != reflect.TypeOf(s) {
		return ErrInvalidLocation("location is not Valid")
	}
	return nil
}

func (msc MsgRegisterVpnService) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.From}
}

type MsgRegisterMasterNode struct {
	Address sdk.AccAddress
}

func NewMsgRegisterMasterNode(addr sdk.AccAddress) MsgRegisterMasterNode {
	return MsgRegisterMasterNode{
		Address: addr,
	}

}
func (msc MsgRegisterMasterNode) Type() string {
	return "sentinel"
}

func (msc MsgRegisterMasterNode) GetSignBytes() []byte {
	byte_format, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byte_format
}

func (msc MsgRegisterMasterNode) ValidateBasic() sdk.Error {
	if msc.Address == nil {
		return sdk.ErrInvalidAddress("Address type is Invalid")
	}
	return nil
}
func (msc MsgRegisterMasterNode) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.Address}
}

func (msg MsgRegisterMasterNode) Tags() sdk.Tags {
	return sdk.NewTags("Master node address ", []byte(msg.Address.String()))
	// AppendTag("receiver", []byte(msg.Receiver.String()))
}

//
//
//
//
//
type MsgDeleteVpnUser struct {
	From  sdk.AccAddress
	Vaddr sdk.AccAddress
}

func NewMsgDeleteVpnUser(From sdk.AccAddress, Vaddr sdk.AccAddress) MsgDeleteVpnUser {
	return MsgDeleteVpnUser{
		From:  From,
		Vaddr: Vaddr,
	}
}

func (msc MsgDeleteVpnUser) Type() string {
	return "sentinel"
}

func (msc MsgDeleteVpnUser) GetSignBytes() []byte {
	byte_format, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byte_format
}

func (msc MsgDeleteVpnUser) ValidateBasic() sdk.Error {
	if msc.From == nil {
		return sdk.ErrInvalidAddress("Address type is Invalid")
	}
	if msc.Vaddr == nil {
		return sdk.ErrInvalidAddress("VPN Address type is Invalid")
	}
	return nil
}
func (msc MsgDeleteVpnUser) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.From}
}

//
//
//
//
//
type MsgDeleteMasterNode struct {
	Address sdk.AccAddress
	Maddr   sdk.AccAddress
}

func NewMsgDeleteMasterNode(addr sdk.AccAddress, Maddr sdk.AccAddress) MsgDeleteMasterNode {
	return MsgDeleteMasterNode{
		Address: addr,
		Maddr:   Maddr,
	}
}
func (msc MsgDeleteMasterNode) Type() string {
	return "sentinel"
}

func (msc MsgDeleteMasterNode) GetSignBytes() []byte {
	byte_format, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byte_format
}

func (msc MsgDeleteMasterNode) ValidateBasic() sdk.Error {
	if msc.Address == nil {
		return sdk.ErrInvalidAddress("Address type is Invalid")
	}
	if msc.Maddr == nil {
		return sdk.ErrInvalidAddress("VPN Address type is Invalid")
	}
	return nil
}
func (msc MsgDeleteMasterNode) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.Address}
}

//
//
//
//
//
type MsgPayVpnService struct {
	Coins   sdk.Coins
	Vpnaddr sdk.AccAddress
	From    sdk.AccAddress
	Pubkey  crypto.PubKey
}

func NewMsgPayVpnService(coins sdk.Coins, vaddr sdk.AccAddress, from sdk.AccAddress, pubkey crypto.PubKey) MsgPayVpnService {
	return MsgPayVpnService{
		Coins:   coins,
		Vpnaddr: vaddr,
		From:    from,
		Pubkey:  pubkey,
	}

}

func (msc MsgPayVpnService) Type() string {
	return "sentinel"
}
func (msc MsgPayVpnService) GetSignBytes() []byte {
	byte_format, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byte_format
}

func (msc MsgPayVpnService) ValidateBasic() sdk.Error {
	if msc.Coins.IsZero() || !(msc.Coins.IsNotNegative()) {
		return sdk.ErrInsufficientFunds("Error insufficient coins")
	}
	if msc.From == nil || msc.Vpnaddr == nil {
		return sdk.ErrInvalidAddress("Invalid address type")
	}
	return nil
}
func (msc MsgPayVpnService) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.From}
}

//
//
//
//
//
type MsgSigntoVpn struct {
	Coins     sdk.Coins
	Address   sdk.AccAddress
	Sessionid []byte

	From sdk.AccAddress
}

func (msc MsgSigntoVpn) Type() string {
	return "sentinel"
}

func (msc MsgSigntoVpn) GetSignBytes() []byte {
	byte_format, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byte_format
}

func (msc MsgSigntoVpn) ValidateBasic() sdk.Error {
	var a []byte
	if msc.Coins.IsZero() || !(msc.Coins.IsNotNegative()) {
		return sdk.ErrInsufficientFunds("Error insufficent coins")
	}
	if reflect.TypeOf(msc.Sessionid) != reflect.TypeOf(a) {
		return ErrInvalidSessionid(" Invalid SessionId")
	}
	if msc.Address == nil {
		return sdk.ErrInvalidAddress("Invalid Address")
	}
	if msc.From == nil {
		return sdk.ErrInvalidAddress("Invalid  from Address")
	}
	return nil
}
func (msc MsgSigntoVpn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.From}
}

//
//
//
//

type MsgGetVpnPayment struct {
	Coins     sdk.Coins
	Sessionid []byte
	Counter   int64
	Signature crypto.Signature
	From      sdk.AccAddress
	IsFinal   bool
}

func NewMsgGetVpnPayment(coin sdk.Coins, sid []byte, counter int64, addr sdk.AccAddress, sign crypto.Signature, isfinal bool) MsgGetVpnPayment {
	return MsgGetVpnPayment{
		Coins:     coin,
		Sessionid: sid,
		Counter:   counter,
		Signature: sign,
		From:      addr,
		IsFinal:   isfinal,
	}

}
func (msc MsgGetVpnPayment) Type() string {
	return "sentinel"
}

type sign struct {
	Coins     sdk.Coins
	Sessionid []byte
	Counter   int64
	IsFinal   bool
}

func NewSign(coins sdk.Coins, Sess []byte, counter int64, isFinal bool) sign {
	return sign{
		Coins:     coins,
		Sessionid: Sess,
		Counter:   counter,
		IsFinal:   isFinal,
	}
}

func (msc MsgGetVpnPayment) GetSignBytes() []byte {
	byte_format, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byte_format
}

func (msc MsgGetVpnPayment) ValidateBasic() sdk.Error {
	if msc.Coins.IsZero() || !(msc.Coins.IsNotNegative()) {
		return sdk.ErrInsufficientFunds("Error insufficent coins")
	}
	return nil
}
func (msc MsgGetVpnPayment) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.From}
}

//
//
//
//
//
type MsgRefund struct {
	From      sdk.AccAddress
	Sessionid []byte
	//Time      int64
}

func NewMsgRefund(addr sdk.AccAddress, sid []byte) MsgRefund {
	return MsgRefund{
		From:      addr,
		Sessionid: sid,
	}

}
func (msc MsgRefund) Type() string {
	return "sentinel"
}

func (msc MsgRefund) GetSignBytes() []byte {
	byte_format, err := json.Marshal(msc)
	if err != nil {
		return nil
	}
	return byte_format
}

func (msc MsgRefund) ValidateBasic() sdk.Error {
	if msc.Sessionid == nil {
		return ErrInvalidSessionid("SessionId is Invalid")
	}
	return nil
}
func (msc MsgRefund) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msc.From}
}

//msg Decoder fo the Query:
