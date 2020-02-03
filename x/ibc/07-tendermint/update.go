package tendermint

import (
	"errors"
	"time"

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
// - header valset commit verification fails
//
// Tendermint client validity checking uses the bisection algorithm described
// in the [Tendermint spec](https://github.com/tendermint/spec/blob/master/spec/consensus/light-client.md).
func CheckValidityAndUpdateState(
	clientState clientexported.ClientState, header clientexported.Header, chainID string,
	currentTimestamp time.Time,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	tmClientState, ok := clientState.(ClientState)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidClientType, "light client is not from Tendermint",
		)
	}

	tmHeader, ok := header.(Header)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader, "header is not from Tendermint",
		)
	}

	if err := checkValidity(tmClientState, tmHeader, chainID, currentTimestamp); err != nil {
		return nil, nil, err
	}

	tmClientState, consensusState := update(tmClientState, tmHeader)
	return tmClientState, consensusState, nil
}

// checkValidity checks if the Tendermint header is valid.
//
// CONTRACT: assumes header.Height > consensusState.Height
func checkValidity(
	clientState ClientState, header Header, chainID string, currentTimestamp time.Time,
) error {
	// assert trusting period has not yet passed
	if currentTimestamp.Sub(clientState.LatestTimestamp) >= clientState.TrustingPeriod {
		return errors.New("trusting period since last client timestamp already passed")
	}

	// assert header timestamp is not in the future (& transitively that is not past the trusting period)
	if header.Time.Unix() > currentTimestamp.Unix() {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader,
			"header blocktime can't be in the future",
		)
	}

	// assert header timestamp is past current timestamp
	if header.Time.Unix() <= clientState.LatestTimestamp.Unix() {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header blocktime ≤ latest client state block time (%s ≤ %s)",
			header.Time.String(), clientState.LatestTimestamp.String(),
		)
	}

	// assert header height is newer than any we know
	if header.GetHeight() <= clientState.LatestHeight {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header height ≤ latest client state height (%d ≤ %d)", header.GetHeight(), clientState.LatestHeight,
		)
	}

	// basic consistency check
	if err := header.ValidateBasic(chainID); err != nil {
		return err
	}

	return header.ValidatorSet.VerifyCommit(header.ChainID, header.Commit.BlockID, header.Height, header.Commit)
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
