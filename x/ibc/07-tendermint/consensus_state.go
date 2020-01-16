package tendermint

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ConsensusState defines a Tendermint consensus state
type ConsensusState struct {
	Root             commitment.RootI `json:"root" yaml:"root"`
	ValidatorSetHash []byte           `json:"validator_set_hash" yaml:"validator_set_hash"`
}

// ClientType returns Tendermint
func (ConsensusState) ClientType() exported.ClientType {
	return exported.Tendermint
}

// GetRoot returns the commitment Root for the specific
func (cs ConsensusState) GetRoot() commitment.RootI {
	return cs.Root
}

// ValidateBasic
func (cs ConsensusState) ValidateBasic() error {
	if cs.Root == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "root cannot be nil")
	}
	if len(cs.ValidatorSetHash) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "validator set hash cannot be empty")
	}
	return nil
}
