package types

import (
	"bytes"
	"time"

	"github.com/tendermint/tendermint/light"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// CheckHeaderAndUpdateState checks if the provided header is valid, and if valid it will:
// create the consensus state for the header.Height
// and update the client state if the header height is greater than the latest client state height
// It returns an error if:
// - the client or header provided are not parseable to tendermint types
// - the header is invalid
// - header height is less than or equal to the consensus state height
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
func (cs ClientState) CheckHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	header clientexported.Header,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	tmHeader, ok := header.(Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "expected type %T, got %T", Header{}, header,
		)
	}

	// Get consensus bytes from clientStore
	tmConsState, err := GetConsensusState(clientStore, cdc, tmHeader.TrustedHeight)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(
			err, "could not get consensus state from clientstore at TrustedHeight: %d", tmHeader.TrustedHeight,
		)
	}

	if err := checkValidity(&cs, tmConsState, tmHeader, ctx.BlockTime()); err != nil {
		return nil, nil, err
	}

	newClientState, consensusState := update(&cs, tmHeader)
	return newClientState, consensusState, nil
}

// checkTrustedHeader checks that consensus state matches trusted fields of Header
func checkTrustedHeader(header Header, consState *ConsensusState) error {
	if header.TrustedHeight != consState.Height {
		return sdkerrors.Wrapf(
			ErrInvalidHeaderHeight,
			"trusted header height %d does not match consensus state height %d",
			header.TrustedHeight, consState.Height,
		)
	}
	// assert that trustedVals is NextValidators of last trusted header
	// to do this, we check that trustedVals.Hash() == consState.NextValidatorsHash
	tvalHash := header.TrustedValidators.Hash()
	if !bytes.Equal(consState.NextValidatorsHash, tvalHash) {
		return sdkerrors.Wrapf(
			ErrInvalidValidatorSet,
			"trusted validators %s, does not hash to latest trusted validators. Expected: %X, got: %X",
			header.TrustedValidators, consState.NextValidatorsHash, tvalHash,
		)
	}
	return nil
}

// checkValidity checks if the Tendermint header is valid.
// CONTRACT: consState.Height == header.TrustedHeight
func checkValidity(
	clientState *ClientState, consState *ConsensusState,
	header Header, currentTimestamp time.Time,
) error {
	if err := checkTrustedHeader(header, consState); err != nil {
		return err
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

	// Verify next header with the passed-in trustedVals
	// - asserts trusting period not passed
	// - assert header timestamp is not past the trusting period
	// - assert header timestamp is past latest stored consensus state timestamp
	// - assert that a TrustLevel proportion of TrustedValidators signed new Commit
	err := light.Verify(
		clientState.GetChainID(), &signedHeader,
		header.TrustedValidators, &header.SignedHeader, header.ValidatorSet,
		clientState.TrustingPeriod, currentTimestamp, clientState.MaxClockDrift, clientState.TrustLevel.ToTendermint(),
	)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to verify header")
	}
	return nil
}

// update the consensus state from a new header
func update(clientState *ClientState, header Header) (*ClientState, *ConsensusState) {
	if uint64(header.Height) > clientState.LatestHeight {
		clientState.LatestHeight = uint64(header.Height)
	}
	consensusState := &ConsensusState{
		Height:             uint64(header.Height),
		Timestamp:          header.Time,
		Root:               commitmenttypes.NewMerkleRoot(header.AppHash),
		NextValidatorsHash: header.NextValidatorsHash,
	}

	return clientState, consensusState
}
