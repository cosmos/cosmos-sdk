package types

import (
	"time"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// NewConsensusState creates a new ConsensusState instance.
func NewConsensusState(
	timestamp time.Time, root commitmenttypes.MerkleRoot, height uint64,
	nextValsHash tmbytes.HexBytes,
) ConsensusState {
	return ConsensusState{
		Timestamp:          timestamp,
		Root:               root,
		Height:             height,
		NextValidatorsHash: nextValsHash,
	}
}

// ClientType returns Tendermint
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// GetRoot returns the commitment Root for the specific
func (cs ConsensusState) GetRoot() commitmentexported.Root {
	return cs.Root
}

// GetHeight returns the height for the specific consensus state
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Height
}

// GetTimestamp returns block time in nanoseconds at which the consensus state was stored
func (cs ConsensusState) GetTimestamp() uint64 {
	return uint64(cs.Timestamp.UnixNano())
}

// ValidateBasic defines a basic validation for the tendermint consensus state.
func (cs ConsensusState) ValidateBasic() error {
	if cs.Root.Empty() {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "root cannot be empty")
	}
	if err := tmtypes.ValidateHash(cs.NextValidatorsHash); err != nil {
		return sdkerrors.Wrap(err, "next validators hash is invalid")
	}
	if cs.Height == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "height cannot be 0")
	}
	if cs.Timestamp.IsZero() {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp cannot be zero Unix time")
	}
	if cs.Timestamp.UnixNano() < 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp cannot be negative Unix time")
	}
	return nil
}
