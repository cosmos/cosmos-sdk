package covenant

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

func (mcc MsgCreateCovenant) Type() string {
	return "covenant"
}

func (mcc MsgCreateCovenant) GetSignBytes() []byte {
	b, _ := json.Marshal(mcc)
	return b
}

func (mcc MsgCreateCovenant) ValidateBasic() sdk.Error {
	return nil
}

func (mcc MsgCreateCovenant) GetSigners() []sdk.Address {
	return []sdk.Address{mcc.Sender}
}

type MsgSettleCovenant struct {
	CovID    int64       `json:"covid"`
	Settler  sdk.Address `json:"settler"`
	Receiver sdk.Address `json:"receiver"`
}

func (msc MsgSettleCovenant) Type() string {
	return "covenant"
}

func (msc MsgSettleCovenant) GetSignBytes() []byte {
	b, _ := json.Marshal(msc)
	return b
}

func (msc MsgSettleCovenant) ValidateBasic() sdk.Error {
	return nil
}

func (msc MsgSettleCovenant) GetSigners() []sdk.Address {
	return []sdk.Address{msc.Settler}
}
