package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// VerifyClientConsensusState verifies a proof of the consensus state of the
// specified client stored on the target machine.
func (k Keeper) VerifyClientConsensusState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	consensusHeight uint64,
	proof commitmentexported.Proof,
	consensusState clientexported.ConsensusState,
) (err error) {
	clientID := connection.GetClientID()
	clientState, found := k.clientKeeper.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	targetConsState, found := k.clientKeeper.GetClientConsensusState(ctx, clientID, height)
	if !found {
		return sdkerrors.Wrapf(clienttypes.ErrConsensusStateNotFound, "clientID: %s with height: %d", clientID, height)
	}

	clientState, err = clientState.VerifyClientConsensusState(
		k.aminoCdc, targetConsState.GetRoot(), height, connection.GetCounterparty().GetClientID(), consensusHeight, connection.GetCounterparty().GetPrefix(), proof, consensusState,
	)

	if err == nil {
		k.clientKeeper.SetClientState(ctx, clientState)
	}

	return err
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (k Keeper) VerifyConnectionState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof commitmentexported.Proof,
	connectionID string,
	connectionEnd exported.ConnectionI, // opposite connection
) (err error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	clientState, err = clientState.VerifyConnectionState(
		k.cdc, height, connection.GetCounterparty().GetPrefix(), proof, connectionID, connectionEnd, consensusState,
	)

	if err != nil {
		return err
	}

	k.clientKeeper.SetClientState(ctx, clientState)
	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	channel channelexported.ChannelI,
) (err error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	clientState, err = clientState.VerifyChannelState(
		k.cdc, height, connection.GetCounterparty().GetPrefix(), proof,
		portID, channelID, channel, consensusState,
	)

	if err != nil {
		return err
	}

	k.clientKeeper.SetClientState(ctx, clientState)
	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketCommitment(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
) (err error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	clientState, err = clientState.VerifyPacketCommitment(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, commitmentBytes, consensusState,
	)

	if err != nil {
		return err
	}

	k.clientKeeper.SetClientState(ctx, clientState)
	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) (err error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	clientState, err = clientState.VerifyPacketAcknowledgement(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, acknowledgement, consensusState,
	)

	if err != nil {
		return err
	}

	k.clientKeeper.SetClientState(ctx, clientState)
	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketAcknowledgementAbsence(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	sequence uint64,
) (err error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	clientState, err = clientState.VerifyPacketAcknowledgementAbsence(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, consensusState,
	)

	if err != nil {
		return err
	}

	k.clientKeeper.SetClientState(ctx, clientState)
	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof commitmentexported.Proof,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) (err error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	clientState, err = clientState.VerifyNextSequenceRecv(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		nextSequenceRecv, consensusState,
	)

	if err != nil {
		return err
	}

	k.clientKeeper.SetClientState(ctx, clientState)
	return nil
}
