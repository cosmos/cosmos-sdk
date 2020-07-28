package types

import (
	"time"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ConsensusState defines a Tendermint consensus state
type ConsensusState struct {
	Timestamp          time.Time               `json:"timestamp" yaml:"timestamp"`
	Root               commitmentexported.Root `json:"root" yaml:"root"`
	Height             uint64                  `json:"height" yaml:"height"`
	NextValidatorsHash tmbytes.HexBytes        `json:"next_validators_hash"` // validators hash for the next block
	ValidatorSet       *tmtypes.ValidatorSet   `json:"validator_set" yaml:"validator_set"`
}

// NewConsensusState creates a new ConsensusState instance.
func NewConsensusState(
	timestamp time.Time, root commitmentexported.Root, height uint64,
	nextValsHash tmbytes.HexBytes, valset *tmtypes.ValidatorSet,
) ConsensusState {
	return ConsensusState{
		Timestamp:          timestamp,
		Root:               root,
		Height:             height,
		NextValidatorsHash: nextValsHash,
		ValidatorSet:       valset,
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
	if cs.Root == nil || cs.Root.Empty() {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "root cannot be empty")
	}
	if cs.ValidatorSet == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "validator set cannot be nil")
	}
	if err := tmtypes.ValidateHash(cs.NextValidatorsHash); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "next validators hash is invalid: %v", err)
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
