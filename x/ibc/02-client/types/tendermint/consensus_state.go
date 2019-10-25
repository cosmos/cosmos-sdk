package tendermint

import (
	"bytes"
	"errors"

	lerr "github.com/tendermint/tendermint/lite/errors"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

var _ exported.ConsensusState = ConsensusState{}

// ConsensusState defines a Tendermint consensus state
type ConsensusState struct {
	ChainID          string                `json:"chain_id" yaml:"chain_id"`
	Height           uint64                `json:"height" yaml:"height"` // NOTE: defined as 'sequence' in the spec
	Root             commitment.RootI      `json:"root" yaml:"root"`
	NextValidatorSet *tmtypes.ValidatorSet `json:"next_validator_set" yaml:"next_validator_set"` // contains the PublicKey
}

// ClientType returns Tendermint
func (ConsensusState) ClientType() exported.ClientType {
	return exported.Tendermint
}

// GetHeight returns the ConsensusState height
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Height
}

// GetRoot returns the commitment Root for the specific
func (cs ConsensusState) GetRoot() commitment.RootI {
	return cs.Root
}

// CheckValidityAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropriate
func (cs ConsensusState) CheckValidityAndUpdateState(header exported.Header) (exported.ConsensusState, error) {
	tmHeader, ok := header.(Header)
	if !ok {
		return nil, errors.New("header not a valid tendermint header")
	}

	if err := cs.checkValidity(tmHeader); err != nil {
		return nil, err
	}

	return cs.update(tmHeader), nil
}

// checkValidity checks if the Tendermint header is valid
//
// CONTRACT: assumes header.Height > consensusState.Height
func (cs ConsensusState) checkValidity(header Header) error {
	// check if the hash from the consensus set and header
	// matches
	nextHash := cs.NextValidatorSet.Hash()
	if cs.Height == uint64(header.Height-1) &&
		!bytes.Equal(nextHash, header.ValidatorsHash) {
		return lerr.ErrUnexpectedValidators(nextHash, header.ValidatorsHash)
	}

	// validate the next validator set hash from the header
	nextHash = header.NextValidatorSet.Hash()
	if !bytes.Equal(header.NextValidatorsHash, nextHash) {
		return lerr.ErrUnexpectedValidators(header.NextValidatorsHash, nextHash)
	}

	// basic consistency check
	if err := header.ValidateBasic(cs.ChainID); err != nil {
		return err
	}

	// abortTransactionUnless(consensusState.publicKey.verify(header.signature))
	return cs.NextValidatorSet.VerifyFutureCommit(
		header.ValidatorSet, cs.ChainID, header.Commit.BlockID, header.Height, header.Commit,
	)
}

// update the consensus state from a new header
func (cs ConsensusState) update(header Header) ConsensusState {
	cs.Height = header.GetHeight()
	cs.Root = commitment.NewRoot(header.AppHash)
	cs.NextValidatorSet = header.NextValidatorSet
	return cs
}
