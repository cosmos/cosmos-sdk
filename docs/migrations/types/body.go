package types

import "github.com/cosmos/cosmos-sdk/types"

const Mnemonic = "parrot mercy coach forget else spike hip impulse metal ask ginger favorite general diary truth bacon involve mass wear rapid duty match share episode"

type BankSendBody struct {
	AccountNumber uint64           `json:"accountNumber"`
	Sequence      uint64           `json:"sequence"`
	Sender        types.AccAddress `json:"sender"`
	Receiver      types.AccAddress `json:"receiver"`

	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`

	ChainID       string  `json:"chainId"`
	Memo          string  `json:"memo,omitempty"`
	Fee           int64   `json:"fees,omitempty"`
	GasAdjustment float64 `json:"gasAdjustment,omitempty"`
	Gas           uint64  `json:"gas"`
}
