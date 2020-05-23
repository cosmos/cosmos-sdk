package keeper

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// ConnOpenInit initialises a connection attempt on chain A.
//
// NOTE: Identifiers are checked on msg validation.
func (k Keeper) ConnOpenInit(
	ctx sdk.Context,
	connectionID, // identifier
	clientID string,
	counterparty types.Counterparty, // desiredCounterpartyConnectionIdentifier, counterpartyPrefix, counterpartyClientIdentifier
) error {
	_, found := k.GetConnection(ctx, connectionID)
	if found {
		return sdkerrors.Wrap(types.ErrConnectionExists, "cannot initialize connection")
	}

	// connection defines chain A's ConnectionEnd
	connection := types.NewConnectionEnd(types.INIT, connectionID, clientID, counterparty, types.GetCompatibleVersions())
	k.SetConnection(ctx, connectionID, connection)

	if err := k.addConnectionToClient(ctx, clientID, connectionID); err != nil {
		return sdkerrors.Wrap(err, "cannot initialize connection")
	}

	k.Logger(ctx).Info(fmt.Sprintf("connection %s state updated: NONE -> INIT", connectionID))
	return nil
}

// ConnOpenTry relays notice of a connection attempt on chain A to chain B (this
// code is executed on chain B).
//
// NOTE:
//  - Here chain A acts as the counterparty
//  - Identifiers are checked on msg validation
func (k Keeper) ConnOpenTry(
	ctx sdk.Context,
	connectionID string, // desiredIdentifier
	counterparty types.Counterparty, // counterpartyConnectionIdentifier, counterpartyPrefix and counterpartyClientIdentifier
	clientID string, // clientID of chainA
	counterpartyVersions []string, // supported versions of chain A
	proofInit commitmentexported.Proof, // proof that chainA stored connectionEnd in state (on ConnOpenInit)
	proofConsensus commitmentexported.Proof, // proof that chainA stored chainB's consensus state at consensus height
	proofHeight uint64, // height at which relayer constructs proof of A storing connectionEnd in state
	consensusHeight uint64, // latest height of chain B which chain A has stored in its chain B client
) error {
	if consensusHeight > uint64(ctx.BlockHeight()) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "invalid consensus height")
	}

	expectedConsensusState, found := k.clientKeeper.GetSelfConsensusState(ctx, consensusHeight)
	if !found {
		return clienttypes.ErrSelfConsensusStateNotFound
	}

	// expectedConnection defines Chain A's ConnectionEnd
	// NOTE: chain A's counterparty is chain B (i.e where this code is executed)
	counterpartyPrefix, err := counterparty.GetPrefix()
	if err != nil {
		return err
	}

	prefix := k.GetCommitmentPrefix(counterpartyPrefix.GetCommitmentType())
	expectedCounterparty, err := types.NewCounterparty(clientID, connectionID, commitmenttypes.NewMerklePrefix(prefix.Bytes()))
	if err != nil {
		return err
	}

	expectedConnection := types.NewConnectionEnd(types.INIT, counterparty.ConnectionID, counterparty.ClientID, expectedCounterparty, counterpartyVersions)

	// chain B picks a version from Chain A's available versions that is compatible
	// with the supported IBC versions
	version := types.PickVersion(counterpartyVersions, types.GetCompatibleVersions())

	// connection defines chain B's ConnectionEnd
	connection := types.NewConnectionEnd(types.UNINITIALIZED, connectionID, clientID, counterparty, []string{version})

	// Check that ChainA committed expectedConnectionEnd to its state
	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofInit, counterparty.ConnectionID,
		expectedConnection,
	); err != nil {
		return err
	}

	// Check that ChainA stored the correct ConsensusState of chainB at the given consensusHeight
	if err := k.VerifyClientConsensusState(
		ctx, connection, proofHeight, consensusHeight, proofConsensus, expectedConsensusState,
	); err != nil {
		return err
	}

	// If connection already exists for connectionID, ensure that the existing connection's counterparty
	// is chainA and connection is on INIT stage
	// Check that existing connection version is on desired version of current handshake
	previousConnection, found := k.GetConnection(ctx, connectionID)
	if found {
		prevConnCountPrefix, err := previousConnection.Counterparty.GetPrefix()
		if err != nil {
			return err
		}

		if !(previousConnection.State == types.INIT &&
			previousConnection.Counterparty.ConnectionID == counterparty.ConnectionID &&
			bytes.Equal(prevConnCountPrefix.Bytes(), counterpartyPrefix.Bytes()) &&
			previousConnection.ClientID == clientID &&
			previousConnection.Counterparty.ClientID == counterparty.ClientID &&
			previousConnection.Versions[0] == version) {
			return sdkerrors.Wrap(types.ErrInvalidConnection, "cannot relay connection attempt")
		}
	}

	// Set connection state to TRYOPEN and store in chainB state
	connection.State = types.TRYOPEN
	if err := k.addConnectionToClient(ctx, clientID, connectionID); err != nil {
		return sdkerrors.Wrap(err, "cannot relay connection attempt")
	}

	k.SetConnection(ctx, connectionID, connection)
	k.Logger(ctx).Info(fmt.Sprintf("connection %s state updated: NONE -> TRYOPEN ", connectionID))
	return nil
}

// ConnOpenAck relays acceptance of a connection open attempt from chain B back
// to chain A (this code is executed on chain A).
//
// NOTE: Identifiers are checked on msg validation.
func (k Keeper) ConnOpenAck(
	ctx sdk.Context,
	connectionID string,
	version string, // version that ChainB chose in ConnOpenTry
	proofTry commitmentexported.Proof, // proof that connectionEnd was added to ChainB state in ConnOpenTry
	proofConsensus commitmentexported.Proof, // proof that chainB has stored ConsensusState of chainA on its client
	proofHeight uint64, // height that relayer constructed proofTry
	consensusHeight uint64, // latest height of chainA that chainB has stored on its chainA client
) error {
	// Check that chainB client hasn't stored invalid height
	if consensusHeight > uint64(ctx.BlockHeight()) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidHeight, "invalid consensus height")
	}

	// Retrieve connection
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound, "cannot relay ACK of open attempt")
	}

	// Check connection on ChainA is on correct state: INIT or TRYOPEN
	if connection.State != types.INIT && connection.State != types.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is not INIT (got %s)", connection.State.String(),
		)
	}

	// Check that ChainB's proposed version is one of chainA's accepted versions
	if types.LatestVersion(connection.Versions) != version {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidVersion,
			"connection version does't match provided one (%s ≠ %s)", types.LatestVersion(connection.Versions), version,
		)
	}

	// Retrieve chainA's consensus state at consensusheight
	expectedConsensusState, found := k.clientKeeper.GetSelfConsensusState(ctx, consensusHeight)
	if !found {
		return clienttypes.ErrSelfConsensusStateNotFound
	}

	counterpartyPrefix, err := connection.Counterparty.GetPrefix()
	if err != nil {
		return err
	}

	prefix := k.GetCommitmentPrefix(counterpartyPrefix.GetCommitmentType())
	expectedCounterparty, err := types.NewCounterparty(connection.ClientID, connectionID, commitmenttypes.NewMerklePrefix(prefix.Bytes()))
	if err != nil {
		return err
	}

	expectedConnection := types.NewConnectionEnd(types.TRYOPEN, connection.Counterparty.ConnectionID, connection.Counterparty.ClientID, expectedCounterparty, []string{version})

	// Ensure that ChainB stored expected connectionEnd in its state during ConnOpenTry
	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofTry, connection.Counterparty.ConnectionID,
		expectedConnection,
	); err != nil {
		return err
	}

	// Ensure that ChainB has stored the correct ConsensusState for chainA at the consensusHeight
	if err := k.VerifyClientConsensusState(
		ctx, connection, proofHeight, consensusHeight, proofConsensus, expectedConsensusState,
	); err != nil {
		return err
	}

	// Update connection state to Open
	connection.State = types.OPEN
	connection.Versions = []string{version}
	k.SetConnection(ctx, connectionID, connection)
	k.Logger(ctx).Info(fmt.Sprintf("connection %s state updated: INIT -> OPEN ", connectionID))
	return nil
}

// ConnOpenConfirm confirms opening of a connection on chain A to chain B, after
// which the connection is open on both chains (this code is executed on chain B).
//
// NOTE: Identifiers are checked on msg validation.
func (k Keeper) ConnOpenConfirm(
	ctx sdk.Context,
	connectionID string,
	proofAck commitmentexported.Proof, // proof that connection opened on ChainA during ConnOpenAck
	proofHeight uint64, // height that relayer constructed proofAck
) error {
	// Retrieve connection
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound, "cannot relay ACK of open attempt")
	}

	// Check that connection state on ChainB is on state: TRYOPEN
	if connection.State != types.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is not TRYOPEN (got %s)", connection.State.String(),
		)
	}

	counterpartyPrefix, err := connection.Counterparty.GetPrefix()
	if err != nil {
		return err
	}

	prefix := k.GetCommitmentPrefix(counterpartyPrefix.GetCommitmentType())
	expectedCounterparty, err := types.NewCounterparty(connection.ClientID, connectionID, commitmenttypes.NewMerklePrefix(prefix.Bytes()))
	if err != nil {
		return err
	}
	expectedConnection := types.NewConnectionEnd(types.OPEN, connection.Counterparty.ConnectionID, connection.Counterparty.ClientID, expectedCounterparty, connection.Versions)

	// Check that connection on ChainA is open
	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofAck, connection.Counterparty.ConnectionID,
		expectedConnection,
	); err != nil {
		return err
	}

	// Update ChainB's connection to Open
	connection.State = types.OPEN
	k.SetConnection(ctx, connectionID, connection)
	k.Logger(ctx).Info(fmt.Sprintf("connection %s state updated: TRYOPEN -> OPEN ", connectionID))
	return nil
}
