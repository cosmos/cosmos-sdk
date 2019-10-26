package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TransferPacketData defines a struct for the packet payload
type TransferPacketData struct {
	Denomination string         `json:"denomination" yaml:"denomination"`
	Amount       sdk.Int        `json:"amount" yaml:"amount"`
	Sender       sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver     string         `json:"receiver" yaml:"receiver"`
	Source       bool           `json:"source" yaml:"source"`
}

func (tpd TransferPacketData) MarshalAmino() ([]byte, error) {
	return MouduleCdc.MarshalBinaryBare(tpd)
}

func (tpd *TransferPacketData) UnmarshalAmino(bz []byte) (err error) {
	return MouduleCdc.UnmarshalBinaryBare(bz, tpd)
}

func (tpd TransferPacketData) Marshal() []byte {
	return MouduleCdc.MustMarshalBinaryBare(tpd)
}

func (tpd TransferPacketData) MarshalJSON() ([]byte, error) {
	return MouduleCdc.MarshalJSON(tpd)
}

func (tpd *TransferPacketData) UnmarshalJSON(bz []byte) (err error) {
	return MouduleCdc.UnmarshalJSON(bz, tpd)
}

func (tpd TransferPacketData) String() string {
	return fmt.Sprintf(`TransferPacketData:
	Denomination          %s
	Amount:               %s
	Sender:               %s
	Receiver:             %s
	Source:               %v`,
		tpd.Denomination,
		tpd.Amount.String(),
		tpd.Sender.String(),
		tpd.Receiver,
		tpd.Source,
	)
}

func (tpd TransferPacketData) Validate() error {
	if !tpd.Amount.IsPositive() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAmount, "invalid amount")
	}

	if tpd.Sender.Empty() || len(tpd.Receiver) == 0 {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid address")
	}

	return nil
}
