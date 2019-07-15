package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type MsgOpenInit struct {
	ConnectionID       string
	Connection         Connection
	CounterpartyClient string
	NextTimeout        uint64
	Signer             sdk.AccAddress
}

var _ sdk.Msg = MsgOpenInit{}

func (msg MsgOpenInit) Route() string {
	return "ibc"
}

func (msg MsgOpenInit) Type() string {
	return "open-init"
}

func (msg MsgOpenInit) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenInit) GetSignBytes() []byte {
	return nil // TODO
}

func (msg MsgOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenTry struct {
	ConnectionID       string
	Connection         Connection
	CounterpartyClient string
	Timeout            uint64
	NextTimeout        uint64
	Proofs             []commitment.Proof
	Signer             sdk.AccAddress
}

var _ sdk.Msg = MsgOpenTry{}

type MsgOpenAck struct {
	ConnectionID string
	Timeout      uint64
	NextTimeout  uint64
	Proofs       []commitment.Proof
	Signer       sdk.AccAddress
}

var _ sdk.Msg = MsgOpenAck{}

type MsgOpenConfirm struct {
	ConnectionID string
	Timeout      uint64
	Proofs       []commitment.Proof
	Signer       sdk.AccAddress
}

var _ sdk.Msg = MsgOpenConfirm{}
