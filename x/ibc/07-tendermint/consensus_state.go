package tendermint

import (
	"bytes"
	"fmt"

	lerr "github.com/tendermint/tendermint/lite/errors"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienterrors "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

var _ exported.ConsensusState = ConsensusState{}

// ConsensusState defines a Tendermint consensus state
type ConsensusState struct {
	ChainID          string                `json:"chain_id" yaml:"chain_id"`
	Height           uint64                `json:"height" yaml:"height"` // NOTE: defined as 'sequence' in the spec
	Root             commitment.RootI      `json:"root" yaml:"root"`
	ValidatorSet     *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
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

// GetCommitter returns the commmitter that committed the ConsensusState
func (cs ConsensusState) GetCommitter() exported.Committer {
	return Committer{
		ValidatorSet:   cs.ValidatorSet,
		Height:         cs.Height,
		NextValSetHash: cs.NextValidatorSet.Hash(),
	}
}

// CheckValidityAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropriate
func (cs ConsensusState) CheckValidityAndUpdateState(header exported.Header) (exported.ConsensusState, error) {
	tmHeader, ok := header.(Header)
	if !ok {
		return nil, sdkerrors.Wrap(
			clienterrors.ErrInvalidHeader, "header is not from Tendermint",
		)
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
	if header.GetHeight() < cs.Height {
		return sdkerrors.Wrap(
			clienterrors.ErrInvalidHeader,
			fmt.Sprintf("header height < consensus height (%d < %d)", header.GetHeight(), cs.Height),
		)
	}

	// basic consistency check
	if err := header.ValidateBasic(cs.ChainID); err != nil {
		return err
	}

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

	// abortTransactionUnless(consensusState.publicKey.verify(header.signature))
	return header.ValidatorSet.VerifyFutureCommit(
		cs.NextValidatorSet, cs.ChainID, header.Commit.BlockID, header.Height, header.Commit,
	)
}

// update the consensus state from a new header
func (cs ConsensusState) update(header Header) ConsensusState {
	cs.Height = header.GetHeight()
	cs.Root = commitment.NewRoot(header.AppHash)
	cs.ValidatorSet = header.ValidatorSet
	cs.NextValidatorSet = header.NextValidatorSet
	return cs
}
