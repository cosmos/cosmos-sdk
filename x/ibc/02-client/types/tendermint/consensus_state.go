package tendermint

import (
	"bytes"
	"errors"

	lerr "github.com/tendermint/tendermint/lite/errors"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

var _ exported.ConsensusState = ConsensusState{}

// ConsensusState defines a Tendermint consensus state
type ConsensusState struct {
	ChainID          string                `json:"chain_id" yaml:"chain_id"`
	Height           uint64                `json:"height" yaml:"height"` // NOTE: defined as 'sequence' in the spec
	Root             ics23.Root            `json:"root" yaml:"root"`
	NextValidatorSet *tmtypes.ValidatorSet `json:"next_validator_set" yaml:"next_validator_set"` // contains the PublicKey
}

// Kind returns Tendermint
func (ConsensusState) Kind() exported.Kind {
	return exported.Tendermint
}

// GetHeight returns the ConsensusState height
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Height
}

// GetRoot returns the commitment Rootgit
func (cs ConsensusState) GetRoot() ics23.Root {
	return cs.Root
}

// CheckValidityAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropiate
func (cs ConsensusState) CheckValidityAndUpdateState(header exported.Header) (exported.ConsensusState, error) {
	tmHeader, ok := header.(Header)
	if !ok {
		return nil, errors.New("header is not from a tendermint consensus")
	}

	if err := cs.checkValidity(tmHeader); err != nil {
		return nil, err
	}

	return cs.update(tmHeader), nil
}

// checkValidity checks if the Tendermint header is valid
func (cs ConsensusState) checkValidity(header Header) error {
	// TODO: shouldn't we check that header.Height > cs.Height?
	nextHash := cs.NextValidatorSet.Hash()
	if cs.Height == uint64(header.Height-1) &&
		!bytes.Equal(header.ValidatorsHash, nextHash) {
		return lerr.ErrUnexpectedValidators(header.ValidatorsHash, nextHash)
	}

	nextHash = header.NextValidatorSet.Hash()
	if !bytes.Equal(header.NextValidatorsHash, nextHash) {
		return lerr.ErrUnexpectedValidators(header.NextValidatorsHash, nextHash)
	}

	if err := header.ValidateBasic(cs.ChainID); err != nil {
		return err
	}

	return cs.NextValidatorSet.VerifyFutureCommit(
		header.ValidatorSet, cs.ChainID, header.Commit.BlockID, header.Height, header.Commit,
	)
}

// update the consensus state from a new header
func (cs ConsensusState) update(header Header) ConsensusState {
	return ConsensusState{
		ChainID:          cs.ChainID,
		Height:           uint64(header.Height),
		Root:             merkle.NewRoot(header.AppHash),
		NextValidatorSet: header.NextValidatorSet,
	}
}
