package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ics02types "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ConnOpenInit initialises a connection attempt on chain A.
func (k Keeper) ConnOpenInit(
	ctx sdk.Context,
	connectionID, // identifier
	clientID string,
	counterparty types.Counterparty, // desiredCounterpartyConnectionIdentifier, counterpartyPrefix, counterpartyClientIdentifier
) error {
	// TODO: validateConnectionIdentifier(identifier)
	_, found := k.GetConnection(ctx, connectionID)
	if found {
		return sdkerrors.Wrap(types.ErrConnectionExists(k.codespace), "cannot initialize connection")
	}

	// connection defines chain A's ConnectionEnd
	connection := types.NewConnectionEnd(clientID, counterparty, k.getCompatibleVersions())
	connection.State = types.INIT

	k.SetConnection(ctx, connectionID, connection)
	err := k.addConnectionToClient(ctx, clientID, connectionID)
	if err != nil {
		sdkerrors.Wrap(err, "cannot initialize connection")
	}

	return nil
}

// ConnOpenTry relays notice of a connection attempt on chain A to chain B (this
// code is executed on chain B).
//
// NOTE: here chain A acts as the counterparty
func (k Keeper) ConnOpenTry(
	ctx sdk.Context,
	connectionID string, // desiredIdentifier
	counterparty types.Counterparty, // counterpartyConnectionIdentifier, counterpartyPrefix and counterpartyClientIdentifier
	clientID string,
	counterpartyVersions []string,
	proofInit ics23.Proof,
	proofHeight uint64,
	consensusHeight uint64,
) error {
	// TODO: validateConnectionIdentifier(identifier)
	if consensusHeight > uint64(ctx.BlockHeight()) {
		return errors.New("invalid consensus height") // TODO: sdk.Error
	}

	expectedConsensusState, found := k.clientKeeper.GetConsensusState(ctx, clientID)
	if !found {
		return errors.New("client consensus state not found") // TODO: use ICS02 error
	}

	// expectedConn defines Chain A's ConnectionEnd
	// NOTE: chain A's counterparty is chain B (i.e where this code is executed)
	// TODO: prefix should be `getCommitmentPrefix()`
	expectedCounterparty := types.NewCounterparty(counterparty.ClientID, connectionID, counterparty.Prefix)
	expectedConn := types.NewConnectionEnd(clientID, expectedCounterparty, counterpartyVersions)
	expectedConn.State = types.INIT

	// chain B picks a version from Chain A's available versions
	version := k.pickVersion(counterpartyVersions)

	// connection defines chain B's ConnectionEnd
	connection := types.NewConnectionEnd(clientID, counterparty, []string{version})

	ok := k.verifyMembership(
		ctx, connection, proofHeight, proofInit,
		types.ConnectionPath(connectionID), expectedConn,
	)
	if !ok {
		return errors.New("couldn't verify connection membership on counterparty's client") // TODO: sdk.Error
	}

	ok = k.verifyMembership(
		ctx, connection, proofHeight, proofInit,
		ics02types.ConsensusStatePath(counterparty.ClientID), expectedConsensusState,
	)
	if !ok {
		return errors.New("couldn't verify consensus state membership on counterparty's client") // TODO: sdk.Error
	}

	_, found = k.GetConnection(ctx, connectionID)
	if found {
		return sdkerrors.Wrap(types.ErrConnectionExists(k.codespace), "cannot relay connection attempt")
	}

	if !checkVersion(version, counterpartyVersions[0]) {
		return errors.New("versions don't match") // TODO: sdk.Error
	}

	connection.State = types.TRYOPEN
	err := k.addConnectionToClient(ctx, clientID, connectionID)
	if err != nil {
		return sdkerrors.Wrap(err, "cannot relay connection attempt")
	}

	k.SetConnection(ctx, connectionID, connection)
	return nil
}

// ConnOpenAck relays acceptance of a connection open attempt from chain B back
// to chain A (this code is executed on chain A).
func (k Keeper) ConnOpenAck(
	ctx sdk.Context,
	connectionID string,
	version string,
	proofTry ics23.Proof,
	proofHeight uint64,
	consensusHeight uint64,
) error {
	// TODO: validateConnectionIdentifier(identifier)
	if consensusHeight > uint64(ctx.BlockHeight()) {
		return errors.New("invalid consensus height") // TODO: sdk.Error
	}

	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound(k.codespace), "cannot relay ACK of open attempt")
	}

	if connection.State != types.INIT {
		return errors.New("connection is in a non valid state") // TODO: sdk.Error
	}

	if !checkVersion(connection.Versions[0], version) {
		return errors.New("versions don't match") // TODO: sdk.Error
	}

	expectedConsensusState, found := k.clientKeeper.GetConsensusState(ctx, connection.ClientID)
	if !found {
		return errors.New("client consensus state not found") // TODO: use ICS02 error
	}

	prefix := getCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientID, connectionID, prefix)
	expectedConn := types.NewConnectionEnd(connection.ClientID, expectedCounterparty, []string{version})
	expectedConn.State = types.TRYOPEN

	ok := k.verifyMembership(
		ctx, connection, proofHeight, proofTry,
		types.ConnectionPath(connection.Counterparty.ConnectionID), expectedConn,
	)
	if !ok {
		return errors.New("couldn't verify connection membership on counterparty's client") // TODO: sdk.Error
	}

	ok = k.verifyMembership(
		ctx, connection, proofHeight, proofTry,
		ics02types.ConsensusStatePath(connection.Counterparty.ClientID), expectedConsensusState,
	)
	if !ok {
		return errors.New("couldn't verify consensus state membership on counterparty's client") // TODO: sdk.Error
	}

	connection.State = types.OPEN

	// abort if version is the last one
	// TODO: what does this suppose to mean ?
	compatibleVersions := getCompatibleVersions()
	if compatibleVersions[len(compatibleVersions)-1] == version {
		return errors.New("versions don't match") // TODO: sdk.Error
	}

	connection.Versions = []string{version}
	k.SetConnection(ctx, connectionID, connection)
	return nil
}

// ConnOpenConfirm confirms opening of a connection on chain A to chain B, after
// which the connection is open on both chains (this code is executed on chain B).
func (k Keeper) ConnOpenConfirm(
	ctx sdk.Context,
	connectionID string,
	proofAck ics23.Proof,
	proofHeight uint64,
) error {
	// TODO: validateConnectionIdentifier(identifier)
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound(k.codespace), "cannot relay ACK of open attempt")
	}

	if connection.State != types.TRYOPEN {
		return errors.New("connection is in a non valid state") // TODO: sdk.Error
	}

	prefix := getCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientID, connectionID, prefix)
	expectedConn := types.NewConnectionEnd(connection.ClientID, expectedCounterparty, connection.Versions)
	expectedConn.State = types.OPEN

	ok := k.verifyMembership(
		ctx, connection, proofHeight, proofAck,
		types.ConnectionPath(connection.Counterparty.ConnectionID), expectedConn,
	)
	if !ok {
		return errors.New("couldn't verify connection membership on counterparty's client") // TODO: sdk.Error
	}

	connection.State = types.OPEN
	k.SetConnection(ctx, connectionID, connection)
	return nil
}
