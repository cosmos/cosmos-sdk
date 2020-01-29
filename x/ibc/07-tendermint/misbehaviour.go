package tendermint

import (
	"bytes"
	"errors"
	"time"

	lite "github.com/tendermint/tendermint/lite2"

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
	trustingPeriod time.Duration,
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

	if err := checkMisbehaviour(
		tmClientState, tmConsensusState, tmEvidence, height, trustingPeriod,
	); err != nil {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	tmClientState.FrozenHeight = uint64(tmEvidence.GetHeight())

	return tmClientState, nil
}

// checkMisbehaviour checks if the evidence provided is a valid light client misbehaviour
func checkMisbehaviour(
	clientState ClientState, consensusState ConsensusState, evidence Evidence,
	height uint64, trustingPeriod time.Duration,
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

	return lite.Verify(
		evidence.ChainID, &evidence.Header1.SignedHeader, evidence.FromValidatorSet,
		&evidence.Header2.SignedHeader, evidence.Header2.ValidatorSet, trustingPeriod,
		time.Now(), lite.DefaultTrustLevel,
	)
}
