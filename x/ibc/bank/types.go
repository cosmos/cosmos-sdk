package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc"
)

type PayloadCoins struct {
	SrcAddr  sdk.AccAddress `json:"src-addr"`
	DestAddr sdk.AccAddress `json:"dest-addr"`
	Coins    sdk.Coins      `json:"coins"`
}

func (p PayloadCoins) Type() string {
	return "ibc/bank"
}

func (p PayloadCoins) ValidateBasic() sdk.Error {
	if !p.Coins.IsValid() {
		return sdk.ErrInvalidCoins(p.Coins.String())
	}
	if !p.Coins.IsPositive() {
		return sdk.ErrInvalidCoins(p.Coins.String())
	}
	return nil
}

func (p PayloadCoins) DatagramType() ibc.DatagramType {
	return ibc.PacketType
}

func (p PayloadCoins) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{p.SrcAddr}
}

type PayloadCoinsFail struct {
	PayloadCoins
}

func (p PayloadCoinsFail) DatagramType() ibc.DatagramType {
	return ibc.ReceiptType
}
