package paymentchannels

import (
	"encoding/json"
	"fmt"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/types"
)

// OpenChannelMsg - high level transaction to create a new channel
type OpenChannelMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

// SettleChannelMsg - high level transaction to submit a settlement
type SettleChannelMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

// CloseChannelMsg - high level transaction to close a channel
type CloseChannelMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

// NewOpenChannelMsg - construct msg to open new channel
func NewOpenChannelMsg(in []Input, out []Output) SendMsg {
	return SendMsg{Inputs: in, Outputs: out}
}

// Type - Implements Msg.
func (msg NewCreateChannelMsg) Type() string { return "paymentchannels/create" } // TODO: "bank/send"

// ValidateBasic - Implements Msg.
func (msg NewCreateChannelMsg) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs()
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs()
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut types.Coins
	for _, in := range msg.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return err
		}
		totalIn = totalIn.Plus(in.Coins)
	}
	for _, out := range msg.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err
		}
		totalOut = totalOut.Plus(out.Coins)
	}
	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInvalidCoins(totalIn.String()) // TODO
	}
	return nil
}

func (msg NewCreateChannelMsg) String() string {
	return fmt.Sprintf("SendMsg{%v->%v}", msg.Inputs, msg.Outputs)
}

// Get - Implements Msg.
func (msg SendMsg) Get(key interface{}) (value interface{}) {
	return nil
}

// GetSignBytes - Implements Msg.
func (msg SendMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// GetSigners - Implements Msg.
func (msg SendMsg) GetSigners() []crypto.Address {
	addrs := make([]crypto.Address, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}
