package keeper

import (
	"bytes"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// ConnOpenInit initialises a connection attempt on chain A. The generated connection identifier
// is returned.
//
// NOTE: Msg validation verifies the supplied identifiers and ensures that the counterparty
// connection identifier is empty.
func (k Keeper) ConnOpenInit(
	ctx sdk.Context,
	clientID string,
	counterparty types.Counterparty, // counterpartyPrefix, counterpartyClientIdentifier
	version *types.Version,
	delayPeriod uint64,
) (string, error) {
	versions := types.GetCompatibleVersions()
	if version != nil {
		if !types.IsSupportedVersion(version) {
			return "", sdkerrors.Wrap(types.ErrInvalidVersion, "version is not supported")
		}

		versions = []exported.Version{version}
	}

	// connection defines chain A's ConnectionEnd
	connectionID := k.GenerateConnectionIdentifier(ctx)
	connection := types.NewConnectionEnd(types.INIT, clientID, counterparty, types.ExportedVersionsToProto(versions), delayPeriod)
	k.SetConnection(ctx, connectionID, connection)

	if err := k.addConnectionToClient(ctx, clientID, connectionID); err != nil {
		return "", err
	}

	k.Logger(ctx).Info("connection state updated", "connection-id", connectionID, "previous-state", "NONE", "new-state", "INIT")

	defer func() {
		telemetry.IncrCounter(1, "ibc", "connection", "open-init")
	}()

	return connectionID, nil
}

// ConnOpenTry relays notice of a connection attempt on chain A to chain B (this
// code is executed on chain B).
//
// NOTE:
//  - Here chain A acts as the counterparty
//  - Identifiers are checked on msg validation
func (k Keeper) ConnOpenTry(
	ctx sdk.Context,
	previousConnectionID string, // previousIdentifier
	counterparty types.Counterparty, // counterpartyConnectionIdentifier, counterpartyPrefix and counterpartyClientIdentifier
	delayPeriod uint64,
	clientID string, // clientID of chainA
	clientState exported.ClientState, // clientState that chainA has for chainB
	counterpartyVersions []exported.Version, // supported versions of chain A
	proofInit []byte, // proof that chainA stored connectionEnd in state (on ConnOpenInit)
	proofClient []byte, // proof that chainA stored a light client of chainB
	proofConsensus []byte, // proof that chainA stored chainB's consensus state at consensus height
	proofHeight exported.Height, // height at which relayer constructs proof of A storing connectionEnd in state
	consensusHeight exported.Height, // latest height of chain B which chain A has stored in its chain B client
) (string, error) {
	var (
		connectionID       string
		previousConnection types.ConnectionEnd
		found              bool
	)

	// empty connection identifier indicates continuing a previous connection handshake
	if previousConnectionID != "" {
		// ensure that the previous connection exists
		previousConnection, found = k.GetConnection(ctx, previousConnectionID)
		if !found {
			return "", sdkerrors.Wrapf(types.ErrConnectionNotFound, "previous connection does not exist for supplied previous connectionID %s", previousConnectionID)
		}

		// ensure that the existing connection's
		// counterparty is chainA and connection is on INIT stage.
		// Check that existing connection versions for initialized connection is equal to compatible
		// versions for this chain.
		// ensure that existing connection's delay period is the same as desired delay period.
		if !(previousConnection.Counterparty.ConnectionId == "" &&
			bytes.Equal(previousConnection.Counterparty.Prefix.Bytes(), counterparty.Prefix.Bytes()) &&
			previousConnection.ClientId == clientID &&
			previousConnection.Counterparty.ClientId == counterparty.ClientId &&
			previousConnection.DelayPeriod == delayPeriod) {
			return "", sdkerrors.Wrap(types.ErrInvalidConnection, "connection fields mismatch previous connection fields")
		}

		if !(previousConnection.State == types.INIT) {
			return "", sdkerrors.Wrapf(types.ErrInvalidConnectionState, "previous connection state is in state %s, expected INIT", previousConnection.State)
		}

		// continue with previous connection
		connectionID = previousConnectionID

	} else {
		// generate a new connection
		connectionID = k.GenerateConnectionIdentifier(ctx)
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	if consensusHeight.GTE(selfHeight) {
		return "", sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"consensus height is greater than or equal to the current block height (%s >= %s)", consensusHeight, selfHeight,
		)
	}

	// validate client parameters of a chainB client stored on chainA
	if err := k.clientKeeper.ValidateSelfClient(ctx, clientState); err != nil {
		return "", err
	}

	expectedConsensusState, found := k.clientKeeper.GetSelfConsensusState(ctx, consensusHeight)
	if !found {
		return "", sdkerrors.Wrap(clienttypes.ErrSelfConsensusStateNotFound, consensusHeight.String())
	}

	// expectedConnection defines Chain A's ConnectionEnd
	// NOTE: chain A's counterparty is chain B (i.e where this code is executed)
	// NOTE: chainA and chainB must have the same delay period
	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(clientID, "", commitmenttypes.NewMerklePrefix(prefix.Bytes()))
	expectedConnection := types.NewConnectionEnd(types.INIT, counterparty.ClientId, expectedCounterparty, types.ExportedVersionsToProto(counterpartyVersions), delayPeriod)

	supportedVersions := types.GetCompatibleVersions()
	if len(previousConnection.Versions) != 0 {
		supportedVersions = previousConnection.GetVersions()
	}

	// chain B picks a version from Chain A's available versions that is compatible
	// with Chain B's supported IBC versions. PickVersion will select the intersection
	// of the supported versions and the counterparty versions.
	version, err := types.PickVersion(supportedVersions, counterpartyVersions)
	if err != nil {
		return "", err
	}

	// connection defines chain B's ConnectionEnd
	connection := types.NewConnectionEnd(types.TRYOPEN, clientID, counterparty, []*types.Version{version}, delayPeriod)

	// Check that ChainA committed expectedConnectionEnd to its state
	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofInit, counterparty.ConnectionId,
		expectedConnection,
	); err != nil {
		return "", err
	}

	// Check that ChainA stored the clientState provided in the msg
	if err := k.VerifyClientState(ctx, connection, proofHeight, proofClient, clientState); err != nil {
		return "", err
	}

	// Check that ChainA stored the correct ConsensusState of chainB at the given consensusHeight
	if err := k.VerifyClientConsensusState(
		ctx, connection, proofHeight, consensusHeight, proofConsensus, expectedConsensusState,
	); err != nil {
		return "", err
	}

	// store connection in chainB state
	if err := k.addConnectionToClient(ctx, clientID, connectionID); err != nil {
		return "", sdkerrors.Wrapf(err, "failed to add connection with ID %s to client with ID %s", connectionID, clientID)
	}

	k.SetConnection(ctx, connectionID, connection)
	k.Logger(ctx).Info("connection state updated", "connection-id", connectionID, "previous-state", previousConnection.State.String(), "new-state", "TRYOPEN")

	defer func() {
		telemetry.IncrCounter(1, "ibc", "connection", "open-try")
	}()

	return connectionID, nil
}

// ConnOpenAck relays acceptance of a connection open attempt from chain B back
// to chain A (this code is executed on chain A).
//
// NOTE: Identifiers are checked on msg validation.
func (k Keeper) ConnOpenAck(
	ctx sdk.Context,
	connectionID string,
	clientState exported.ClientState, // client state for chainA on chainB
	version *types.Version, // version that ChainB chose in ConnOpenTry
	counterpartyConnectionID string,
	proofTry []byte, // proof that connectionEnd was added to ChainB state in ConnOpenTry
	proofClient []byte, // proof of client state on chainB for chainA
	proofConsensus []byte, // proof that chainB has stored ConsensusState of chainA on its client
	proofHeight exported.Height, // height that relayer constructed proofTry
	consensusHeight exported.Height, // latest height of chainA that chainB has stored on its chainA client
) error {
	// Check that chainB client hasn't stored invalid height
	selfHeight := clienttypes.GetSelfHeight(ctx)
	if consensusHeight.GTE(selfHeight) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"consensus height is greater than or equal to the current block height (%s >= %s)", consensusHeight, selfHeight,
		)
	}

	// Retrieve connection
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound, connectionID)
	}

	// Verify the provided version against the previously set connection state
	switch {
	// connection on ChainA must be in INIT or TRYOPEN
	case connection.State != types.INIT && connection.State != types.TRYOPEN:
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is not INIT or TRYOPEN (got %s)", connection.State.String(),
		)

	// if the connection is INIT then the provided version must be supproted
	case connection.State == types.INIT && !types.IsSupportedVersion(version):
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is in INIT but the provided version is not supported %s", version,
		)

	// if the connection is in TRYOPEN then the version must be the only set version in the
	// retreived connection state.
	case connection.State == types.TRYOPEN && (len(connection.Versions) != 1 || !proto.Equal(connection.Versions[0], version)):
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is in TRYOPEN but the provided version (%s) is not set in the previous connection versions %s", version, connection.Versions,
		)
	}

	// validate client parameters of a chainA client stored on chainB
	if err := k.clientKeeper.ValidateSelfClient(ctx, clientState); err != nil {
		return err
	}

	// Retrieve chainA's consensus state at consensusheight
	expectedConsensusState, found := k.clientKeeper.GetSelfConsensusState(ctx, consensusHeight)
	if !found {
		return clienttypes.ErrSelfConsensusStateNotFound
	}

	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientId, connectionID, commitmenttypes.NewMerklePrefix(prefix.Bytes()))
	expectedConnection := types.NewConnectionEnd(types.TRYOPEN, connection.Counterparty.ClientId, expectedCounterparty, []*types.Version{version}, connection.DelayPeriod)

	// Ensure that ChainB stored expected connectionEnd in its state during ConnOpenTry
	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofTry, counterpartyConnectionID,
		expectedConnection,
	); err != nil {
		return err
	}

	// Check that ChainB stored the clientState provided in the msg
	if err := k.VerifyClientState(ctx, connection, proofHeight, proofClient, clientState); err != nil {
		return err
	}

	// Ensure that ChainB has stored the correct ConsensusState for chainA at the consensusHeight
	if err := k.VerifyClientConsensusState(
		ctx, connection, proofHeight, consensusHeight, proofConsensus, expectedConsensusState,
	); err != nil {
		return err
	}

	k.Logger(ctx).Info("connection state updated", "connection-id", connectionID, "previous-state", connection.State.String(), "new-state", "OPEN")

	defer func() {
		telemetry.IncrCounter(1, "ibc", "connection", "open-ack")
	}()

	// Update connection state to Open
	connection.State = types.OPEN
	connection.Versions = []*types.Version{version}
	connection.Counterparty.ConnectionId = counterpartyConnectionID
	k.SetConnection(ctx, connectionID, connection)
	return nil
}

// ConnOpenConfirm confirms opening of a connection on chain A to chain B, after
// which the connection is open on both chains (this code is executed on chain B).
//
// NOTE: Identifiers are checked on msg validation.
func (k Keeper) ConnOpenConfirm(
	ctx sdk.Context,
	connectionID string,
	proofAck []byte, // proof that connection opened on ChainA during ConnOpenAck
	proofHeight exported.Height, // height that relayer constructed proofAck
) error {
	// Retrieve connection
	connection, found := k.GetConnection(ctx, connectionID)
	if !found {
		return sdkerrors.Wrap(types.ErrConnectionNotFound, connectionID)
	}

	// Check that connection state on ChainB is on state: TRYOPEN
	if connection.State != types.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidConnectionState,
			"connection state is not TRYOPEN (got %s)", connection.State.String(),
		)
	}

	prefix := k.GetCommitmentPrefix()
	expectedCounterparty := types.NewCounterparty(connection.ClientId, connectionID, commitmenttypes.NewMerklePrefix(prefix.Bytes()))
	expectedConnection := types.NewConnectionEnd(types.OPEN, connection.Counterparty.ClientId, expectedCounterparty, connection.Versions, connection.DelayPeriod)

	// Check that connection on ChainA is open
	if err := k.VerifyConnectionState(
		ctx, connection, proofHeight, proofAck, connection.Counterparty.ConnectionId,
		expectedConnection,
	); err != nil {
		return err
	}

	// Update ChainB's connection to Open
	connection.State = types.OPEN
	k.SetConnection(ctx, connectionID, connection)
	k.Logger(ctx).Info("connection state updated", "connection-id", connectionID, "previous-state", "TRYOPEN", "new-state", "OPEN")

	defer func() {
		telemetry.IncrCounter(1, "ibc", "connection", "open-confirm")
	}()

	return nil
}
