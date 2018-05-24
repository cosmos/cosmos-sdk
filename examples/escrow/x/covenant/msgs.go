package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgCreateCovenant struct {
	Sender    sdk.Address   `json:"sender"`
	Settlers  []sdk.Address `json:"settlers"`
	Receivers []sdk.Address `json:"receivers"`
	Amount    sdk.Coins     `json:"amount"`
}

func (ccm MsgCreateCovenant) Type() string {
	return "covenant"
}

func (ccm MsgCreateCovenant) GetSignBytes() []byte {
	b, _ := json.Marshal(ccm)
	return b
}

func (ccm MsgCreateCovenant) ValidateBasic() sdk.Error {
	return nil
}

func (ccm MsgCreateCovenant) GetSigners() []sdk.Address {
	return []sdk.Address{ccm.Sender}
}

type MsgSettleCovenant struct {
	CovID    int64       `json:"covid"`
	Settler  sdk.Address `json:"settler"`
	Receiver sdk.Address `json:"receiver"`
}

func (scm MsgSettleCovenant) Type() string {
	return "covenant"
}

func (scm MsgSettleCovenant) GetSignBytes() []byte {
	b, _ := json.Marshal(scm)
	return b
}

func (scm MsgSettleCovenant) ValidateBasic() sdk.Error {
	return nil
}

func (scm MsgSettleCovenant) GetSigners() []sdk.Address {
	return []sdk.Address{scm.Settler}
}
