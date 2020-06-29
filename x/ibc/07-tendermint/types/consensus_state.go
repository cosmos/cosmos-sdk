package types

import (
	"time"

	ics23 "github.com/confio/ics23/go"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ConsensusState defines a Tendermint consensus state
type ConsensusState struct {
	Timestamp    time.Time               `json:"timestamp" yaml:"timestamp"`
	Root         commitmentexported.Root `json:"root" yaml:"root"`
	Height       uint64                  `json:"height" yaml:"height"`
	ValidatorSet *tmtypes.ValidatorSet   `json:"validator_set" yaml:"validator_set"`

	// Optionally include params
	UnbondingPeriod time.Duration         `json:"unbonding_period" yaml:"unbonding_period"`
	ProofSpecs      []*ics23.ProofSpec    `json:"proof_specs" yaml:"proof_specs"`
	ConsensusParams *abci.ConsensusParams `json:"consensus_params" yaml:"consensus_params"`
}

// NewConsensusState creates a new ConsensusState instance.
func NewConsensusState(
	timestamp time.Time, root commitmentexported.Root, height uint64,
	valset *tmtypes.ValidatorSet,
) ConsensusState {
	return ConsensusState{
		Timestamp:    timestamp,
		Root:         root,
		Height:       height,
		ValidatorSet: valset,
	}
}

// NewConsensusStateWithParams creates a new ConsensusState instance
// with the given paramaterss for client verification
func NewConsensusStateWithParams(
	timestamp time.Time, root commitmentexported.Root, height uint64, valset *tmtypes.ValidatorSet,
	ubdPeriod time.Duration, proofSpecs []*ics23.ProofSpec, consensusParams *abci.ConsensusParams,
) ConsensusState {
	return ConsensusState{
		Timestamp:       timestamp,
		Root:            root,
		Height:          height,
		ValidatorSet:    valset,
		UnbondingPeriod: ubdPeriod,
		ProofSpecs:      proofSpecs,
		ConsensusParams: consensusParams,
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
	if cs.Height == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "height cannot be 0")
	}
	if cs.Timestamp.IsZero() {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp cannot be zero Unix time")
	}
	if cs.Timestamp.UnixNano() < 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp cannot be negative Unix time")
	}
	if cs.ProofSpecs != nil {
		for i, ps := range cs.ProofSpecs {
			if ps == nil {
				return sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "ProofSpec is nil in ProofSpecs list at position %d", i)
			}
		}
	}
	return nil
}
