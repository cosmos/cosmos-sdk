package types

import (
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

// NewPacketData contructs a new PacketData
func NewPacketData(amount sdk.Coins, sender, receiver sdk.AccAddress, source bool) PacketData {
	return PacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
	}
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
	if !pd.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("transfer amount must be positive")
	}
	if !pd.Amount.IsValid() {
		return sdk.ErrInvalidCoins("transfer amount is invalid")
	}
	if pd.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if pd.Receiver.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
	}
	return nil
}
