package tendermint

import (
	"time"

	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// CheckValidityAndUpdateState checks if the provided header is valid, and if valid it will:
// create the consensus state for the header.Height
// and updated the client state if appropriate
// It returns an error if:
// - the client or header provided are not parseable to tendermint types
// - the header is invalid
// - header height is lower than consensusstate height
// - header valset commit verification fails
// - header timestamp is past the trusting period in relation to the consensus state
// - header timestamp is less than or equal to the consensus state timestamp
//
// UpdateClient may be used to either create a consensus state for:
// - a future height greater than the latest client state height
// - a past height that was skipped during bisection
// If we are updating to a past height, a consensus state is created for that height to be persisted in client store
// If we are updating to a future height, the consensus state is created and the client state is updated to reflect
// the new latest height
// Tendermint client validity checking uses the bisection algorithm described
// in the [Tendermint spec](https://github.com/tendermint/spec/blob/master/spec/consensus/light-client.md).
func CheckValidityAndUpdateState(
	clientState clientexported.ClientState, consState clientexported.ConsensusState,
	header clientexported.Header, currentTimestamp time.Time,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	tmClientState, ok := clientState.(types.ClientState)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidClientType, "expected type %T, got %T", types.ClientState{}, clientState,
		)
	}

	tmConsState, ok := consState.(types.ConsensusState)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidConsensus, "expected type %T, got %T", types.ConsensusState{}, consState,
		)
	}

	tmHeader, ok := header.(types.Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "expected type %T, got %T", types.Header{}, header,
		)
	}

	if err := checkValidity(tmClientState, tmConsState, tmHeader, currentTimestamp); err != nil {
		return nil, nil, err
	}

	tmClientState, consensusState := update(tmClientState, tmHeader)
	return tmClientState, consensusState, nil
}

// checkValidity checks if the Tendermint header is valid.
func checkValidity(
	clientState types.ClientState, consState types.ConsensusState, header types.Header, currentTimestamp time.Time,
) error {
	// assert trusting period has not yet passed
	if currentTimestamp.Sub(consState.Timestamp) >= clientState.TrustingPeriod {
		return sdkerrors.Wrapf(
			types.ErrTrustingPeriodExpired,
			"current timestamp minus the consensus state timestamp is greater than or equal to the trusting period (%s >= %s)",
			currentTimestamp.Sub(consState.Timestamp), clientState.TrustingPeriod,
		)
	}

	// assert header timestamp is not past the trusting period
	if header.Time.Sub(consState.Timestamp) >= clientState.TrustingPeriod {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader,
			"header blocktime is outside trusting period from last client timestamp",
		)
	}

	// assert header timestamp is past latest consensus state timestamp
	if header.Time.Unix() <= consState.Timestamp.Unix() {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header blocktime ≤ consensus state block time (%s ≤ %s)",
			header.Time.UTC(), consState.Timestamp.UTC(),
		)
	}

	// assert header height is newer than consensus state
	if header.GetHeight() <= consState.Height {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header height ≤ consensus state height (%d ≤ %d)", header.GetHeight(), consState.Height,
		)
	}

	// Construct a trusted header using the fields in consensus state
	// Only Height, Time, and NextValidatorsHash are necessary for verification
	trustedHeader := tmtypes.Header{
		Height:             int64(consState.Height),
		Time:               consState.Timestamp,
		NextValidatorsHash: consState.NextValidatorsHash,
	}
	signedHeader := tmtypes.SignedHeader{
		Header: &trustedHeader,
	}

	// Verify next header with the last header's validatorset as trusted validatorset
	err := lite.Verify(
		clientState.GetChainID(), &signedHeader,
		consState.ValidatorSet, &header.SignedHeader, header.ValidatorSet,
		clientState.TrustingPeriod, currentTimestamp, clientState.MaxClockDrift, clientState.TrustLevel,
	)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to verify header")
	}
	return nil
}

// update the consensus state from a new header
func update(clientState types.ClientState, header types.Header) (types.ClientState, types.ConsensusState) {
	if uint64(header.Height) > clientState.LatestHeight {
		clientState.LatestHeight = uint64(header.Height)
	}
	consensusState := types.ConsensusState{
		Height:             uint64(header.Height),
		Timestamp:          header.Time,
		Root:               commitmenttypes.NewMerkleRoot(header.AppHash),
		NextValidatorsHash: header.NextValidatorsHash,
		ValidatorSet:       header.ValidatorSet,
	}

	return clientState, consensusState
}
