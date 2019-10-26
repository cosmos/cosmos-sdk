package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TransferPacketData defines a struct for the packet payload
type TransferPacketData struct {
	Denomination string  `json:"denomination" yaml:"denomination"`
	Amount       sdk.Int `json:"amount" yaml:"amount"`
	Sender       string  `json:"sender" yaml:"sender"`
	Receiver     string  `json:"receiver" yaml:"receiver"`
	Source       bool    `json:"source" yaml:"source"`
}

func (tpd TransferPacketData) MarshalAmino() ([]byte, error) {
	return ModuleCdc.MarshalBinaryBare(tpd)
}

func (tpd *TransferPacketData) UnmarshalAmino(bz []byte) (err error) {
	return ModuleCdc.UnmarshalBinaryBare(bz, tpd)
}

func (tpd TransferPacketData) Marshal() []byte {
	return ModuleCdc.MustMarshalBinaryBare(tpd)
}

type TransferPacketDataAlias TransferPacketData

// MarshalJSON implements the json.Marshaler interface.
func (tpd TransferPacketData) MarshalJSON() ([]byte, error) {
	return json.Marshal((TransferPacketDataAlias)(tpd))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (tpd *TransferPacketData) UnmarshalJSON(bz []byte) (err error) {
	return json.Unmarshal(bz, (*TransferPacketDataAlias)(tpd))
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
		tpd.Sender,
		tpd.Receiver,
		tpd.Source,
	)
}

func (tpd TransferPacketData) Validate() error {
	if !tpd.Amount.IsPositive() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAmount, "invalid amount")
	}

	if len(tpd.Sender) == 0 || len(tpd.Receiver) == 0 {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid address")
	}

	return nil
}
