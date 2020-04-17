package localhost

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ConsensusState defines a Localhost consensus state. It is defined as an empty struct.
type ConsensusState struct{}

// ClientType returns Localhost
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.Localhost
}

// GetRoot returns a nil root
func (cs ConsensusState) GetRoot() commitmentexported.Root {
	return nil
}

// GetHeight returns the 0
func (cs ConsensusState) GetHeight() uint64 {
	return 0
}

// ValidateBasic defines a basic validation for the localhost consensus state.
func (cs ConsensusState) ValidateBasic() error {
	return nil
}
