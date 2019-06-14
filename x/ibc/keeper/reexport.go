package ibc

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

type (
	ConsensusState = client.ConsensusState
	Packet         = channel.Packet
)
