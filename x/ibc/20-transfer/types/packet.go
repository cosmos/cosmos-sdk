package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PacketDataTransfer defines a struct for the packet payload
type PacketDataTransfer struct {
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`     // the tokens to be transferred
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`     // the sender address
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"` // the recipient address on the destination chain
	Source   bool           `json:"source" yaml:"source"`     // indicates if the sending chain is the source chain of the tokens to be transferred
}

func (pt PacketDataTransfer) MarshalAmino() ([]byte, error) {
	return ModuleCdc.MarshalBinaryBare(pt)
}

func (pt *PacketDataTransfer) UnmarshalAmino(bz []byte) (err error) {
	return ModuleCdc.UnmarshalBinaryBare(bz, pt)
}

func (pt PacketDataTransfer) Marshal() []byte {
	return ModuleCdc.MustMarshalBinaryBare(pt)
}

type PacketDataTransferAlias PacketDataTransfer

// MarshalJSON implements the json.Marshaler interface.
func (pt PacketDataTransfer) MarshalJSON() ([]byte, error) {
	return json.Marshal((PacketDataTransferAlias)(pt))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (pt *PacketDataTransfer) UnmarshalJSON(bz []byte) (err error) {
	return json.Unmarshal(bz, (*PacketDataTransferAlias)(pt))
}

func (pt PacketDataTransfer) String() string {
	return fmt.Sprintf(`PacketDataTransfer:
	Amount:               %s
	Sender:               %s
	Receiver:             %s
	Source:               %v`,
		pt.Amount.String(),
		pt.Sender,
		pt.Receiver,
		pt.Source,
	)
}

// ValidateBasic performs a basic check of the packet fields
func (pt PacketDataTransfer) ValidateBasic() sdk.Error {
	if !pt.Amount.IsValid() {
		return sdk.ErrInvalidCoins("transfer amount is invalid")
	}
	if !pt.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("transfer amount must be positive")
	}
	if pt.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if pt.Receiver.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
	}
	return nil
}
