package types

import (
	"github.com/tendermint/tendermint/crypto"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

var _ clientexported.ConsensusState = ConsensusState{}

// ConsensusState defines a Solo Machine consensus state
type ConsensusState struct {
	Sequence uint64 `json:"sequence" yaml:"sequence"`

	PublicKey crypto.PubKey `json:"public_key" yaml: "public_key"`
}

// ClientType returns Solo Machine type
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight returns the sequence number
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Sequence
}

// GetRoot returns nil as solo machines do not have roots
func (cs ConsensusState) GetRoot() commitmentexported.Root {
	return nil
}

// ValidateBasic defines basic validation for the solo machine consensus state.
func (cs ConsensusState) ValidateBasic() error {
	if cs.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "sequence cannot be 0")
	}

	if cs.PublicKey == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "public key cannot be empty")
	}

	return nil
}
