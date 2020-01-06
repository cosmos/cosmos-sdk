package tendermint

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
)

// CheckMisbehaviourAndUpdateState
func CheckMisbehaviourAndUpdateState(
	clientState clientexported.ClientState, committer clientexported.Committer,
	misbehaviour clientexported.Misbehaviour,
) (clientexported.ClientState, error) {

	tmClientState, ok := clientState.(ClientState)
	if !ok {
		return nil, sdkerrors.Wrap(errors.ErrInvalidClientType, "client state type is not Tendermint")
	}

	tmCommitter, ok := committer.(Committer)
	if !ok {
		return nil, sdkerrors.Wrap(errors.ErrInvalidClientType, "committer type is not Tendermint")
	}

	tmEvidence, ok := misbehaviour.(Evidence)
	if !ok {
		return nil, sdkerrors.Wrap(errors.ErrInvalidClientType, "committer type is not Tendermint")
	}

	if err := CheckMisbehaviour(tmCommitter, tmEvidence); err != nil {
		return nil, sdkerrors.Wrap(errors.ErrInvalidEvidence, err.Error())
	}

	tmClientState.FrozenHeight = uint64(tmEvidence.GetHeight())

	return clientState, nil
}

// CheckMisbehaviour checks if the evidence provided is a valid light client misbehaviour
func CheckMisbehaviour(trustedCommitter Committer, evidence Evidence) error {
	if err := evidence.ValidateBasic(); err != nil {
		return err
	}

	trustedValSet := trustedCommitter.ValidatorSet

	// Evidence is within trusting period. ValidatorSet must have 2/3 similarity with trustedCommitter ValidatorSet
	// check that the validator sets on both headers are valid given the last trusted validatorset
	// less than or equal to evidence height
	if err := trustedValSet.VerifyFutureCommit(
		evidence.Header1.ValidatorSet, evidence.ChainID,
		evidence.Header1.Commit.BlockID, evidence.Header1.Height, evidence.Header1.Commit,
	); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidEvidence, "validator set in header 1 has too much change from last known committer: %v", err)
	}
	if err := trustedValSet.VerifyFutureCommit(
		evidence.Header2.ValidatorSet, evidence.ChainID,
		evidence.Header2.Commit.BlockID, evidence.Header2.Height, evidence.Header2.Commit,
	); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidEvidence, "validator set in header 2 has too much change from last known committer: %v", err)
	}

	return nil
}
