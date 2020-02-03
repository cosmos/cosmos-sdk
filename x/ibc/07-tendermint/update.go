package tendermint

import (
	"errors"
	"time"

	lite "github.com/tendermint/tendermint/lite2"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// CheckValidityAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropriate. It returns an error if:
// - the client or header provided are not parseable to tendermint types
// - the header is invalid
// - header height is lower than the latest client height
// - light client header verification fails
//
// Tendermint client validity checking uses the bisection algorithm described
// in the [Tendermint spec](https://github.com/tendermint/spec/blob/master/spec/consensus/light-client.md).
func CheckValidityAndUpdateState(
	clientState clientexported.ClientState,
	oldHeader, newHeader clientexported.Header,
	chainID string,
	currentTimestamp time.Time,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	tmClientState, ok := clientState.(ClientState)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidClientType, "light client is not from Tendermint",
		)
	}

	tmHeader1, ok := oldHeader.(Header)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader, "header is not from Tendermint",
		)
	}

	tmHeader2, ok := newHeader.(Header)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader, "header is not from Tendermint",
		)
	}

	if err := checkValidity(tmClientState, tmHeader1, tmHeader2, chainID, currentTimestamp); err != nil {
		return nil, nil, err
	}

	tmClientState, consensusState := update(tmClientState, tmHeader2)
	return tmClientState, consensusState, nil
}

// checkValidity checks if the Tendermint header is valid
// NOTE: old and new header validation is checked by lite.Verify
func checkValidity(
	clientState ClientState,
	oldHeader,
	newHeader Header,
	chainID string,
	currentTimestamp time.Time,
) error {
	// assert trusting period has not yet passed
	if currentTimestamp.Sub(clientState.LatestTimestamp) >= clientState.TrustingPeriod {
		return errors.New("trusting period since last client timestamp already passed")
	}

	// assert header timestamp is not in the future (& transitively that is not past the trusting period)
	if newHeader.Time.Unix() > currentTimestamp.Unix() {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader,
			"header blocktime can't be in the future",
		)
	}

	// assert header timestamp is past current timestamp
	if newHeader.Time.Unix() <= clientState.LatestTimestamp.Unix() {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header blocktime ≤ latest client state block time (%s ≤ %s)",
			newHeader.Time.String(), clientState.LatestTimestamp.String(),
		)
	}

	// assert header height is newer than any we know
	if newHeader.GetHeight() <= clientState.LatestHeight {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header height ≤ latest client state height (%d ≤ %d)", newHeader.GetHeight(), clientState.LatestHeight,
		)
	}

	// call tendermint light client verification function
	return lite.Verify(
		chainID, &oldHeader.SignedHeader, oldHeader.NextValidatorSet, &newHeader.SignedHeader,
		newHeader.ValidatorSet, clientState.TrustingPeriod, time.Now(), lite.DefaultTrustLevel,
	)
}

// update the consensus state from a new header
func update(clientState ClientState, header Header) (ClientState, ConsensusState) {
	clientState.LatestHeight = header.GetHeight()
	consensusState := ConsensusState{
		Timestamp:    header.Time,
		Root:         commitment.NewRoot(header.AppHash),
		ValidatorSet: header.ValidatorSet,
	}

	return clientState, consensusState
}
