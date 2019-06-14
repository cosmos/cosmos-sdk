package ibc

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type (
	Proof          = commitment.Proof
	ConsensusState = client.ConsensusState
	Packet         = channel.Packet
)
