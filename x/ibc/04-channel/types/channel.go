package types

// ChannelOrder defines if a channel is ORDERED or UNORDERED
type ChannelOrder byte

// channel order types
const (
	UNORDERED ChannelOrder = iota // packets can be delivered in any order, which may differ from the order in which they were sent.
	ORDERED                       // packets are delivered exactly in the order which they were sent
)

// ChannelState defines if a channel is in one of the following states:
// CLOSED, INIT, OPENTRY or OPEN
type ChannelState byte

// channel state types
const (
	CLOSED  ChannelState = iota // A channel end has been closed and can no longer be used to send or receive packets.
	INIT                        // A channel end has just started the opening handshake.
	OPENTRY                     // A channel end has acknowledged the handshake step on the counterparty chain.
	OPEN                        // A channel end has completed the handshake and is ready to send and receive packets.
)

type Channel struct {
	State          ChannelState `json:"state" yaml:"state"`
	Ordering       ChannelOrder `json:"ordering" yaml:"ordering"`
	Counterparty   Counterparty `json:"counterparty" yaml:"counterparty"`
	ConnectionHops []string     `json:"connection_hops" yaml:"connection_hops"`
	Version        string       `json:"version" yaml:"version "`
}

// NewChannel creates a new Channel instance
func NewChannel(
	state ChannelState, ordering ChannelOrder, counterparty Counterparty,
	hops []string, version string,
) Channel {
	return Channel{
		State:          state,
		Ordering:       ordering,
		Counterparty:   counterparty,
		ConnectionHops: hops,
		Version:        version,
	}
}

// CounterpartyHops returns the connection hops of the counterparty channel.
// The counterparty hops are stored in the inverse order as the channel's.
func (ch Channel) CounterpartyHops() []string {
	counterPartyHops := make([]string, len(ch.ConnectionHops))
	for i, hop := range ch.ConnectionHops {
		counterPartyHops[len(counterPartyHops)-1-i] = hop
	}
	return counterPartyHops
}

type Counterparty struct {
	PortID    string `json:"port_id" yaml:"port_id"`
	ChannelID string `json:"channel_id" yaml:"channel_id"`
}

// NewCounterparty returns a new Counterparty instance
func NewCounterparty(portID, channelID string) Counterparty {
	return Counterparty{
		PortID:    portID,
		ChannelID: channelID,
	}
}
