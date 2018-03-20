package bank

import (
	"encoding/json"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// move this code to appropriate files
// (./tx.go, ./handler.go)
// after check any conflict

// tx.go

// implements sdk.Msg
type IBCSendMsg struct {
	DestChain string
	SendPayload
}

func (msg IBCSendMsg) Type() string { return "bank" }

func (msg IBCSendMsg) ValidateBasic() sdk.Error {
	return msg.SendPayload.ValidateBasic()
}

func (msg IBCSendMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg IBCSendMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.SendPayload.SrcAddr}
}

func (msg IBCSendMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// implements ibc.Payload
type SendPayload struct {
	SrcAddr  sdk.Address
	DestAddr sdk.Address
	Coins    sdk.Coins
}

func (p SendPayload) Type() string {
	return "bank"
}

func (p SendPayload) ValidateBasic() sdk.Error {
	if !p.Coins.IsValid() {
		return sdk.ErrInvalidCoins(p.Coins.String())
	}
	if !p.Coins.IsPositive() {
		return sdk.ErrInvalidCoins(p.Coins.String())
	}
	return nil
}

// handler.go

func handleIBCSendMsg(ctx sdk.Context, ibcs ibc.Sender, ck CoinKeeper, msg IBCSendMsg) sdk.Result {
	p := msg.SendPayload
	_, err := ck.SubtractCoins(ctx, p.SrcAddr, p.Coins)
	if err != nil {
		return err.Result()
	}
	ibcs.Push(ctx, p, msg.DestChain)
	return sdk.Result{}
}

func NewIBCHandler(ck CoinKeeper) ibc.Handler {
	return func(ctx sdk.Context, p ibc.Payload) sdk.Result {
		switch p := p.(type) {
		case SendPayload:
			return handleTransferMsg(ctx, ck, p)
		default:
			errMsg := "Unrecognized bank Payload type: " + reflect.TypeOf(p).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleTransferMsg(ctx sdk.Context, ck CoinKeeper, p SendPayload) sdk.Result {
	_, err := ck.AddCoins(ctx, p.DestAddr, p.Coins)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{}
}
