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

	PubKey crypto.PubKey `json:"pubkey" yaml:"pubkey"`
}

// ClientType returns Solo Machine type
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight returns the sequence number
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Sequence
}

// GetTimestamp returns zero.
func (cs ConsensusState) GetTimestamp() uint64 {
	return 0
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

	if cs.PubKey == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "public key cannot be empty")
	}

	return nil
}
