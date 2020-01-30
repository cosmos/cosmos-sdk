package tendermint

import (
	"bytes"
	"errors"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// CheckMisbehaviourAndUpdateState determines whether or not two conflicting
// headers at the same height would have convinced the light client.
//
// NOTE: assumes provided height is the height at which the consensusState is
// stored.
func CheckMisbehaviourAndUpdateState(
	clientState clientexported.ClientState,
	consensusState clientexported.ConsensusState,
	misbehaviour clientexported.Misbehaviour,
	height uint64, // height at which the consensus state was loaded
) (clientexported.ClientState, error) {

	// cast the interface to specific types before checking for misbehaviour
	tmClientState, ok := clientState.(ClientState)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "client state type is not Tendermint")
	}

	tmConsensusState, ok := consensusState.(ConsensusState)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "consensus state is not Tendermint")
	}

	tmEvidence, ok := misbehaviour.(Evidence)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "evidence type is not Tendermint")
	}

	if err := checkMisbehaviour(tmClientState, tmConsensusState, tmEvidence, height); err != nil {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	tmClientState.FrozenHeight = uint64(tmEvidence.GetHeight())

	return tmClientState, nil
}

// checkMisbehaviour checks if the evidence provided is a valid light client misbehaviour
func checkMisbehaviour(
	clientState ClientState, consensusState ConsensusState, evidence Evidence, height uint64,
) error {
	// NOTE: header height and commitment root assertions are checked with the
	// evidence and msg ValidateBasic functions at the AnteHandler level.

	// check if provided height matches the headers' height
	if height != uint64(evidence.GetHeight()) {
		return sdkerrors.Wrapf(
			ibctypes.ErrInvalidHeight,
			"height ≠ evidence header height (%d ≠ %d)", height, evidence.GetHeight(),
		)
	}

	if !bytes.Equal(consensusState.ValidatorSetHash, evidence.FromValidatorSet.Hash()) {
		return errors.New(
			"the consensus state's validator set hash doesn't match the evidence's one",
		)
	}

	// Evidence is within the trusting period. ValidatorSet must have 2/3 similarity with trusted FromValidatorSet
	// check that the validator sets on both headers are valid given the last trusted validatorset
	// less than or equal to evidence height
	if err := evidence.FromValidatorSet.VerifyFutureCommit(
		evidence.Header1.ValidatorSet, evidence.ChainID,
		evidence.Header1.Commit.BlockID, evidence.Header1.Height, evidence.Header1.Commit,
	); err != nil {
		return fmt.Errorf("validator set in header 1 has too much change from last known validator set: %v", err)
	}

	if err := evidence.FromValidatorSet.VerifyFutureCommit(
		evidence.Header2.ValidatorSet, evidence.ChainID,
		evidence.Header2.Commit.BlockID, evidence.Header2.Height, evidence.Header2.Commit,
	); err != nil {
		return fmt.Errorf("validator set in header 2 has too much change from last known validator set: %v", err)
	}

	return nil
}
