package types

// query routes supported by the IBC channel Querier
const (
	QueryAllChannels               = "channels"
	QueryChannel                   = "channel"
	QueryConnectionChannels        = "connection-channels"
	QueryChannelClientState        = "channel-client-state"
	QueryPacketCommitments         = "packet-commitments"
	QueryUnrelayedAcknowledgements = "unrelayed-acknowledgements"
	QueryUnrelayedPacketSends      = "unrelayed-packet-sends"
)

// NewChannelResponse creates a new ChannelResponse instance
func NewQueryChannelResponse(channel Channel, proof *merkle.Proof, height int64) QueryChannelResponse {
	return QueryChannelResponse{
		path := commitmenttypes.NewMerklePath(strings.Split(host.ChannelPath(portID, channelID), "/"))
		Channel:     &channel,
		Proof:       commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   &path,
		ProofHeight: uint64(height),
	}
}

// NewQueryPacketResponse creates a new PacketResponswe instance
func NewQueryPacketResponse(
	portID, channelID string, sequence uint64, packet Packet, proof *merkle.Proof, height int64,
) QueryPacketResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.PacketCommitmentPath(portID, channelID, sequence), "/"))
	return QueryPacketResponse{
		Packet:      packet,
		Proof:       commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   &path,
		ProofHeight: uint64(height),
	}
}

