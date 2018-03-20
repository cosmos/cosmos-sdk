package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ------------------------------
// Payload
// Payload defines inter-blockchain message
// that can be proved by light-client protocol

type Payload interface {
	Type() string
	ValidateBasic() sdk.Error
}

// ------------------------------
// Packet

// Packet defines a piece of data that can be send between two separate
// blockchains.
type Packet struct {
	Payload   Payload
	SrcChain  string
	DestChain string
}

func (packet Packet) ValidateBasic() sdk.Error {
	/*
		// commented for testing
		if packet.SrcChain == packet.DestChain {
			return ErrIdenticalChains()
		}
	*/
	return packet.Payload.ValidateBasic()
}
