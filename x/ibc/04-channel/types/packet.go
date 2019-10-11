package types

type Packet struct {
	Seq                  uint64 // number corresponds to the order of sends and receives, where a packet with an earlier sequence number must be sent and received before a packet with a later sequence number.
	Timeout              uint64 // indicates a consensus height on the destination chain after which the packet will no longer be processed, and will instead count as having timed-out.
	SourcePortID         string // identifies the port on the sending chain.
	SourceChannelID      string // identifies the channel end on the sending chain.
	DestinationPortID    string // identifies the port on the receiving chain.
	DestinationChannelID string // identifies the channel end on the receiving chain.
	Data                 []byte // opaque value which can be defined by the application logic of the associated modules.
}

// TODO:
type OpaquePacket Packet
