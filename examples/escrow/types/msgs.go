package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CreateCovenantMessage struct {
	Sender    sdk.Address   `json:"sender"`
	Settlers  []sdk.Address `json:"settlers"`
	Receivers []sdk.Address `json:"receivers"`
	Amount    sdk.Coins     `json:"amount"`
}

func (ccm CreateCovenantMessage) Type() string {
	return "CreateCovenantMessage"
}

func (ccm CreateCovenantMessage) GetSignBytes() []byte {
	b, _ := json.Marshal(ccm)
	return b
}

func (ccm CreateCovenantMessage) ValidateBasic() sdk.Error {
	return nil
}

func (ccm CreateCovenantMessage) GetSigners() []sdk.Address {
	return []sdk.Address{ccm.Sender}
}

type SettleCovenantMessage struct {
	CovID    int64       `json:"covid"`
	Settler  sdk.Address `json:"settler"`
	Receiver sdk.Address `json:"receiver"`
}

func (scm SettleCovenantMessage) Type() string {
	return "SettleCovenantMessage"
}

func (scm SettleCovenantMessage) GetSignBytes() []byte {
	b, _ := json.Marshal(scm)
	return b
}

func (scm SettleCovenantMessage) ValidateBasic() sdk.Error {
	return nil
}

func (scm SettleCovenantMessage) GetSigners() []sdk.Address {
	return []sdk.Address{scm.Settler}
}
