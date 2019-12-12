package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

var _ channeltypes.PacketDataI = PacketDataTransfer{}

// PacketDataTransfer defines a struct for the packet payload
type PacketDataTransfer struct {
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`     // the tokens to be transferred
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`     // the sender address
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"` // the recipient address on the destination chain
	Source   bool           `json:"source" yaml:"source"`     // indicates if the sending chain is the source chain of the tokens to be transferred
	Timeout  uint64         `json:"timeout" yaml:"timeout"`
}

// NewPacketDataTransfer contructs a new PacketDataTransfer
func NewPacketDataTransfer(amount sdk.Coins, sender, receiver sdk.AccAddress, source bool, timeout uint64) PacketDataTransfer {
	return PacketDataTransfer{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
		Timeout:  timeout,
	}
}

func (pd PacketDataTransfer) String() string {
	return fmt.Sprintf(`PacketDataTransfer:
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

// Implements ibc.Packet
// ValidateBasic performs a basic check of the packet fields
func (pd PacketDataTransfer) ValidateBasic() sdk.Error {
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

// Implements ibc.Packet
// TODO: need to be hashed(non-unmarshalable) format
func (pd PacketDataTransfer) GetCommitment() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(pd))
}

// Implements ibc.Packet
func (pd PacketDataTransfer) GetTimeoutHeight() uint64 {
	return pd.Timeout
}

// Implements ibc.Packet
func (pd PacketDataTransfer) Type() string {
	return "ics20/transfer"
}
