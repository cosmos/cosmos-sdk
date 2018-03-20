package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ibc "github.com/cosmos/cosmos-sdk/x/ibc"
)

// move this code to appropriate files
// (./tx.go, ./handler.go)
// after check any conflict

// tx.go

type TransferMsg struct {
	SrcAddr  sdk.Address
	DestAddr sdk.Address
	Coins    sdk.Coins
}

func (msg TransferMsg) Type() string {
	return "bank"
}

func (msg TransferMsg) GetSigners() []sdk.Address {

}

func (msg TransferMsg) ValidateBasic() sdk.Error {
	if !msg.Coins.IsValid() {
		return sdk.ErrInvalidCoins("")
	}
	return nil
}

// handler.go

func NewIBCHandler(ibcm ibc.IBCMapper, ck CoinKeeper) ibc.Handler {
	return func(ctx sdk.Context, msg ibc.Msg) sdk.Result {
		switch msg := msg.(type) {
		case TransferMsg:
			return handleTransferMsg(ctx, ibcm, ck, msg)
		}
	}
}

func handleTransferMsg(ctx sdk.Context, ibcm ibc.IBCMapper, ck CoinKeeper, msg TransferMsg) sdk.Result {

}
