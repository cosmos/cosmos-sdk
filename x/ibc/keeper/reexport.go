package ibc

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type (
	Header         = client.Header
	Proof          = commitment.Proof
	ConsensusState = client.ConsensusState
	Packet         = channel.Packet
	Connection     = connection.Connection
	Channel        = channel.Channel
)
