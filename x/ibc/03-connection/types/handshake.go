package types

// Handshake defines a connection between two chains A and B
type Handshake struct {
	ConnA ConnectionEnd
	ConnB ConnectionEnd
}
