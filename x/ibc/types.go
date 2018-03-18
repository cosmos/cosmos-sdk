package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wire "github.com/cosmos/cosmos-sdk/wire"

	// temporal
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/x/cool"
	"github.com/cosmos/cosmos-sdk/x/bank"
	oldwire "github.com/tendermint/go-wire"
)

// ------------------------------
// IBCPacket

// IBCPacket defines a piece of data that can be send between two separate
// blockchains.
type IBCPacket struct {
	SrcAddr   sdk.Address
	DestAddr  sdk.Address
	Coins     sdk.Coins
	SrcChain  string
	DestChain string
}

func NewIBCPacket(srcAddr sdk.Address, destAddr sdk.Address, coins sdk.Coins,
	srcChain string, destChain string) IBCPacket {

	return IBCPacket{
		SrcAddr:   srcAddr,
		DestAddr:  destAddr,
		Coins:     coins,
		SrcChain:  srcChain,
		DestChain: destChain,
	}
}

func (ibcp IBCPacket) ValidateBasic() sdk.Error {
	if ibcp.SrcChain == ibcp.DestChain {
		return ErrIdenticalChains().Trace("")
	}
	return nil
}

// ----------------------------------
// IBCTransferMsg

// IBCTransferMsg defines how another module can send an IBCPacket.
type IBCTransferMsg struct {
	IBCPacket
}

func (msg IBCTransferMsg) Type() string {
	return "ibc"
}

func (msg IBCTransferMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg IBCTransferMsg) GetSignBytes() []byte {
	cdc := makeCodec()
	bz, err := cdc.MarshalBinary(msg.IBCPacket)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg IBCTransferMsg) ValidateBasic() sdk.Error {
	return msg.IBCPacket.ValidateBasic()
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg IBCTransferMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.SrcAddr}
}

// ----------------------------------
// IBCReceiveMsg

// IBCReceiveMsg defines the message that a relayer uses to post an IBCPacket
// to the destination chain.
type IBCReceiveMsg struct {
	IBCPacket
	Relayer  sdk.Address
	Sequence int64
}

func (msg IBCReceiveMsg) Type() string {
	return "ibc"
}

func (msg IBCReceiveMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg IBCReceiveMsg) GetSignBytes() []byte {
	cdc := newCodec()
	bz, err := cdc.MarshalBinary(msg.IBCPacket)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg IBCReceiveMsg) ValidateBasic() sdk.Error {
	return msg.IBCPacket.ValidateBasic()
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg IBCReceiveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}

// -------------------------
// Helpers

func newCodec() *wire.Codec {
	return wire.NewCodec()
}

func makeCodec() *wire.Codec { // basecoin/app.MakeCodec()
	const (
		msgTypeSend        = 0x1
		msgTypeIssue       = 0x2
		msgTypeQuiz        = 0x3
		msgTypeSetTrend    = 0x4
		msgTypeIBCTransfer = 0x5
		msgTypeIBCReceive  = 0x6
	)

	var _ = oldwire.RegisterInterface(
		struct{ sdk.Msg }{},
		oldwire.ConcreteType{bank.SendMsg{}, msgTypeSend},
		oldwire.ConcreteType{bank.IssueMsg{}, msgTypeIssue},
		oldwire.ConcreteType{cool.QuizMsg{}, msgTypeQuiz},
		oldwire.ConcreteType{cool.SetTrendMsg{}, msgTypeSetTrend},
		oldwire.ConcreteType{IBCTransferMsg{}, msgTypeIBCTransfer},
		oldwire.ConcreteType{IBCReceiveMsg{}, msgTypeIBCReceive},
	)

	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&types.AppAccount{}, accTypeApp},
	)

	cdc := wire.NewCodec()
	return cdc
}
