package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
		return sdkerrors.Wrap(types.ErrConnectionExists(k.codespace, connectionID), "cannot initialize connection")
	}

	// connection defines chain A's ConnectionEnd
	connection := types.NewConnectionEnd(types.INIT, clientID, counterparty, types.GetCompatibleVersions())
	k.SetConnection(ctx, connectionID, connection)

	err := k.addConnectionToClient(ctx, clientID, connectionID)
	if err != nil {
		sdkerrors.Wrap(err, "cannot initialize connection")
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
	// XXX: blocked by #5078
	/*
		if consensusHeight > uint64(ctx.BlockHeight()) {
			return errors.New("invalid consensus height") // TODO: sdk.Error
		}

		expectedConsensusState, found := k.clientKeeper.GetConsensusState(ctx, clientID)
		if !found {
			return errors.New("client consensus state not found") // TODO: use ICS02 error
		}
	*/

	// expectedConn defines Chain A's ConnectionEnd
	// NOTE: chain A's counterparty is chain B (i.e where this code is executed)
	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(clientID, connectionID, prefix)
	expectedConn := types.NewConnectionEnd(types.INIT, counterparty.ClientID, expectedCounterparty, counterpartyVersions)

	// chain B picks a version from Chain A's available versions that is compatible
	// with the supported IBC versions
	version := types.PickVersion(counterpartyVersions, types.GetCompatibleVersions())

	// connection defines chain B's ConnectionEnd
	connection := types.NewConnectionEnd(types.NONE, clientID, counterparty, []string{version})
	expConnBz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedConn)
	if err != nil {
		return err
	}

	ok := k.VerifyMembership(
		ctx, connection, proofHeight, proofInit,
		types.ConnectionPath(counterparty.ConnectionID), expConnBz,
	)
	if !ok {
		fmt.Sprintf("couldn't verify connection membership on counterparty's client\n")
		return errors.New("couldn't verify connection membership on counterparty's client") // TODO: sdk.Error
	}

	// XXX: blocked by #5078
	/*
		expConsStateBz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedConsensusState)
		if err != nil {
			return err
		}

		ok = k.VerifyMembership(
			ctx, connection, proofHeight, proofConsensus,
			clienttypes.ConsensusStatePath(counterparty.ClientID), expConsStateBz,
		)
		if !ok {
			fmt.Sprintf("couldn't verify consensus state membership on counterparty's client\n")
			return errors.New("couldn't verify consensus state membership on counterparty's client") // TODO: sdk.Error
		}

	*/

	_, found := k.GetConnection(ctx, connectionID)
	if found {
		return sdkerrors.Wrap(types.ErrConnectionExists(k.codespace, connectionID), "cannot relay connection attempt")
	}

	connection.State = types.TRYOPEN
	err = k.addConnectionToClient(ctx, clientID, connectionID)
	if err != nil {
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
	// XXX: blocked by #5078
	/*
		if consensusHeight > uint64(ctx.BlockHeight()) {
			return errors.New("invalid consensus height") // TODO: sdk.Error
		}
	*/
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound(k.codespace, connectionID), "cannot relay ACK of open attempt")
	}

	if connection.State != types.INIT {
		return types.ErrInvalidConnectionState(
			k.codespace,
			fmt.Sprintf("connection state is not INIT (got %s)", connection.State.String()),
		)
	}

	if types.LatestVersion(connection.Versions) != version {
		return types.ErrInvalidVersion(
			k.codespace,
			fmt.Sprintf("connection version does't match provided one (%s ≠ %s)", types.LatestVersion(connection.Versions), version),
		)
	}

	// XXX: blocked by #5078
	/*
		expectedConsensusState, found := k.clientKeeper.GetConsensusState(ctx, connection.ClientID)
		if !found {
			return errors.New("client consensus state not found") // TODO: use ICS02 error
		}
	*/
	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientID, connectionID, prefix)
	expectedConn := types.NewConnectionEnd(types.TRYOPEN, connection.Counterparty.ClientID, expectedCounterparty, []string{version})

	expConnBz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedConn)
	if err != nil {
		return err
	}

	ok := k.VerifyMembership(
		ctx, connection, proofHeight, proofTry,
		types.ConnectionPath(connection.Counterparty.ConnectionID), expConnBz,
	)
	if !ok {
		return errors.New("couldn't verify connection membership on counterparty's client") // TODO: sdk.Error
	}

	// XXX: blocked by #5078
	/*
		expConsStateBz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedConsensusState)
		if err != nil {
			return err
		}

		ok = k.VerifyMembership(
			ctx, connection, proofHeight, proofConsensus,
			clienttypes.ConsensusStatePath(connection.Counterparty.ClientID), expConsStateBz,
		)
		if !ok {
			return errors.New("couldn't verify consensus state membership on counterparty's client") // TODO: sdk.Error
		}

	*/
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
	proofAck commitment.ProofI,
	proofHeight uint64,
) error {
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound(k.codespace, connectionID), "cannot relay ACK of open attempt")
	}

	if connection.State != types.TRYOPEN {
		return types.ErrInvalidConnectionState(
			k.codespace,
			fmt.Sprintf("connection state is not TRYOPEN (got %s)", connection.State.String()),
		)
	}

	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientID, connectionID, prefix)
	expectedConn := types.NewConnectionEnd(types.OPEN, connection.Counterparty.ClientID, expectedCounterparty, connection.Versions)

	expConnBz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedConn)
	if err != nil {
		return err
	}

	ok := k.VerifyMembership(
		ctx, connection, proofHeight, proofAck,
		types.ConnectionPath(connection.Counterparty.ConnectionID), expConnBz,
	)
	if !ok {
		return types.ErrInvalidCounterpartyConnection(k.codespace)
	}

	connection.State = types.OPEN
	k.SetConnection(ctx, connectionID, connection)
	k.Logger(ctx).Info(fmt.Sprintf("connection %s state updated: TRYOPEN -> OPEN ", connectionID))
	return nil
}
