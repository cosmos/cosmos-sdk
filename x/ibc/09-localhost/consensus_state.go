package localhost

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ConsensusState defines a Localhost consensus state
type ConsensusState struct {
	Root commitmentexported.Root `json:"root" yaml:"root"`
}

// ClientType returns Localhost
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.Localhost
}

// GetRoot returns the commitment Root for the specific
func (cs ConsensusState) GetRoot() commitmentexported.Root {
	return cs.Root
}

// GetHeight returns the height for the specific consensus state
func (cs ConsensusState) GetHeight() uint64 {
	return 0
}

// ValidateBasic defines a basic validation for the localhost consensus state.
func (cs ConsensusState) ValidateBasic() error {
	return nil
}
