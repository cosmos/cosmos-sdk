package adr036

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ExpectedChainID defines the chain id an off-chain message must have
	ExpectedChainID = ""
	// ExpectedAccountNumber defines the account number an off-chain message must have
	ExpectedAccountNumber = 0
	// ExpectedSequence defines the sequence number an off-chain message must have
	ExpectedSequence = 0
	// ExpectedRoute defines the route to use for sdk.Msg.ExpectedRoute() implementation for offchain messages
	ExpectedRoute = "offchain"
)

type msg interface {
	sdk.Msg
	offchain()
}

func (m *MsgSignArbitraryData) offchain() {}

// NewMsgSignArbitraryData is MsgSignArbitraryData's constructor
func NewMsgSignArbitraryData(signer sdk.AccAddress, data []byte) *MsgSignArbitraryData {
	return &MsgSignArbitraryData{
		Signer: signer.String(),
		Data:   data,
	}
}

func (m *MsgSignArbitraryData) Route() string {
	return ExpectedRoute
}

func (m *MsgSignArbitraryData) Type() string {
	return "MsgSignData"
}

func (m *MsgSignArbitraryData) ValidateBasic() error {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return errors.New("invalid signer")
	}
	if signer.Empty() {
		return errors.New("empty signer")
	}
	if len(m.Data) == 0 {
		return errors.New("empty data")
	}
	return nil
}

func (m *MsgSignArbitraryData) GetSignBytes() []byte {
	return legacy.Cdc.MustMarshalJSON(m)
}

func (m *MsgSignArbitraryData) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}
