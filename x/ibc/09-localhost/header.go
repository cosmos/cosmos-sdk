package localhost

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ clientexported.Header = Header{}

// Header defines the Localhost consensus Header
type Header struct {
}

// ClientType defines that the Header is in loop-back mode.
func (h Header) ClientType() clientexported.ClientType {
	return clientexported.Localhost
}

// ConsensusState returns an empty consensus state.
func (h Header) ConsensusState() ConsensusState {
	return ConsensusState{}
}

// GetHeight returns 0.
func (h Header) GetHeight() uint64 {
	return 0
}
