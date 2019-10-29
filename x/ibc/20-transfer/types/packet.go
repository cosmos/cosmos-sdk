package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PacketData defines a struct for the packet payload
type PacketData struct {
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`     // the tokens to be transferred
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`     // the sender address
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"` // the recipient address on the destination chain
	Source   bool           `json:"source" yaml:"source"`     // indicates if the sending chain is the source chain of the tokens to be transferred
}

func (pd PacketData) MarshalAmino() ([]byte, error) {
	return ModuleCdc.MarshalBinaryBare(pd)
}

func (pd *PacketData) UnmarshalAmino(bz []byte) (err error) {
	return ModuleCdc.UnmarshalBinaryBare(bz, pd)
}

func (pd PacketData) Marshal() []byte {
	return ModuleCdc.MustMarshalBinaryBare(pd)
}

type PacketDataAlias PacketData

// MarshalJSON implements the json.Marshaler interface.
func (pd PacketData) MarshalJSON() ([]byte, error) {
	return json.Marshal((PacketDataAlias)(pd))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (pd *PacketData) UnmarshalJSON(bz []byte) (err error) {
	return json.Unmarshal(bz, (*PacketDataAlias)(pd))
}

func (pd PacketData) String() string {
	return fmt.Sprintf(`PacketData:
	Amount:               %s
	Sender:               %s
	Receiver:             %s
	Source:               %v`,
		pd.Amount.String(),
		pd.Sender,
		pd.Receiver,
		pd.Source,
	)
}

// ValidateBasic performs a basic check of the packet fields
func (pd PacketData) ValidateBasic() sdk.Error {
	if !pd.Amount.IsValid() {
		return sdk.ErrInvalidCoins("transfer amount is invalid")
	}
	if !pd.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("transfer amount must be positive")
	}
	if pd.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if pd.Receiver.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
	}
	return nil
}
