package keeper

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
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
	connection := types.NewConnectionEnd(exported.INIT, clientID, counterparty, types.GetCompatibleVersions())
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
	clientID string,
	counterpartyVersions []string,
	proofInit commitment.ProofI,
	proofConsensus commitment.ProofI,
	proofHeight uint64,
	consensusHeight uint64,
) error {
	if consensusHeight > uint64(ctx.BlockHeight()) {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "invalid consensus height")
	}

	expectedConsensusState, found := k.clientKeeper.GetSelfConsensusState(ctx, consensusHeight)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	// expectedConnection defines Chain A's ConnectionEnd
	// NOTE: chain A's counterparty is chain B (i.e where this code is executed)
	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(clientID, connectionID, prefix)
	expectedConnection := types.NewConnectionEnd(exported.INIT, counterparty.ClientID, expectedCounterparty, counterpartyVersions)

	// chain B picks a version from Chain A's available versions that is compatible
	// with the supported IBC versions
	version := types.PickVersion(counterpartyVersions, types.GetCompatibleVersions())

	// connection defines chain B's ConnectionEnd
	connection := types.NewConnectionEnd(exported.UNINITIALIZED, clientID, counterparty, []string{version})

	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofInit, counterparty.ConnectionID,
		expectedConnection,
	); err != nil {
		return err
	}

	if err := k.VerifyClientConsensusState(
		ctx, connection, consensusHeight, proofConsensus, expectedConsensusState,
	); err != nil {
		return err
	}

	previousConnection, found := k.GetConnection(ctx, connectionID)
	if found && !(previousConnection.State == exported.INIT &&
		previousConnection.Counterparty.ConnectionID == counterparty.ConnectionID &&
		bytes.Equal(previousConnection.Counterparty.Prefix.Bytes(), counterparty.Prefix.Bytes()) &&
		previousConnection.ClientID == clientID &&
		previousConnection.Counterparty.ClientID == counterparty.ClientID &&
		previousConnection.Versions[0] == version) {
		return sdkerrors.Wrap(types.ErrInvalidConnection, "cannot relay connection attempt")
	}

	connection.State = exported.TRYOPEN
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
	version string,
	proofTry commitment.ProofI,
	proofConsensus commitment.ProofI,
	proofHeight uint64,
	consensusHeight uint64,
) error {
	if consensusHeight > uint64(ctx.BlockHeight()) {
		return sdkerrors.Wrap(ibctypes.ErrInvalidHeight, "invalid consensus height")
	}

	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound, "cannot relay ACK of open attempt")
	}

	if connection.State != exported.INIT {
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is not INIT (got %s)", connection.State.String(),
		)
	}

	if types.LatestVersion(connection.Versions) != version {
		return sdkerrors.Wrapf(
			ibctypes.ErrInvalidVersion,
			"connection version does't match provided one (%s ≠ %s)", types.LatestVersion(connection.Versions), version,
		)
	}
	expectedConsensusState, found := k.clientKeeper.GetSelfConsensusState(ctx, consensusHeight)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientID, connectionID, prefix)
	expectedConnection := types.NewConnectionEnd(exported.TRYOPEN, connection.Counterparty.ClientID, expectedCounterparty, []string{version})

	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofTry, connection.Counterparty.ConnectionID,
		expectedConnection,
	); err != nil {
		return err
	}

	if err := k.VerifyClientConsensusState(
		ctx, connection, consensusHeight, proofConsensus, expectedConsensusState,
	); err != nil {
		return err
	}

	connection.State = exported.OPEN
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
	proofAck commitment.ProofI,
	proofHeight uint64,
) error {
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound, "cannot relay ACK of open attempt")
	}

	if connection.State != exported.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is not TRYOPEN (got %s)", connection.State.String(),
		)
	}

	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientID, connectionID, prefix)
	expectedConnection := types.NewConnectionEnd(exported.OPEN, connection.Counterparty.ClientID, expectedCounterparty, connection.Versions)

	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofAck, connection.Counterparty.ConnectionID,
		expectedConnection,
	); err != nil {
		return err
	}

	connection.State = exported.OPEN
	k.SetConnection(ctx, connectionID, connection)
	k.Logger(ctx).Info(fmt.Sprintf("connection %s state updated: TRYOPEN -> OPEN ", connectionID))
	return nil
}
