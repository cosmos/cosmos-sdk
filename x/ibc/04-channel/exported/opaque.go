package exported

// OpaquePacketData is useful as a struct for when a Keeper cannot
// unmarshal an IBC packet.
type OpaquePacketData struct {
	data []byte
}

// NewOpaquePacketData creates a wrapper for rawData.
func NewOpaquePacketData(data []byte) OpaquePacketData {
	return OpaquePacketData{
		data: data,
	}
}

// GetBytes is a helper for serialising
func (opd OpaquePacketData) GetBytes() []byte {
	return opd.data
}
