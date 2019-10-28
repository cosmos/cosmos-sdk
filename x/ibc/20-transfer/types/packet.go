package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TransferPacketData defines a struct for the packet payload
type TransferPacketData struct {
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"`
	Source   bool           `json:"source" yaml:"source"`
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
	Amount:               %s
	Sender:               %s
	Receiver:             %s
	Source:               %v`,
		tpd.Amount.String(),
		tpd.Sender,
		tpd.Receiver,
		tpd.Source,
	)
}

// ValidateBasic performs a basic check of the packet fields
func (tpd TransferPacketData) ValidateBasic() sdk.Error {
	if !tpd.Amount.IsValid() {
		return sdk.ErrInvalidCoins("transfer amount is invalid")
	}
	if !tpd.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("transfer amount must be positive")
	}
	if tpd.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if tpd.Receiver.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
	}
	return nil
}
