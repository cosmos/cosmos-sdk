package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// HANDSHAKE PROCESS IS OMITTED

// ICS 02

func (k Keeper) CreateClient(ctx sdk.Context, id string, cstate client.ConsensusState) error {
	_, err := k.client.Create(ctx, id, cstate)
	return err
}

func (k Keeper) QueryClient(ctx sdk.Context, id string) (client.ConsensusState, error) {
	obj, err := k.client.Query(ctx, id)
	if err != nil {
		return nil, err
	}
	return obj.Value(ctx), nil
}

func (k Keeper) QueryClientFrozen(ctx sdk.Context, id string) (bool, error) {
	obj, err := k.client.Query(ctx, id)
	if err != nil {
		return false, err
	}
	return obj.Frozen(ctx), nil
}

func (k Keeper) UpdateClient(ctx sdk.Context, id string, header client.Header) error {
	obj, err := k.client.Query(ctx, id)
	if err != nil {
		return err
	}

	return obj.Update(ctx, header)
}

func (k Keeper) DeleteClient(ctx sdk.Context, id string) error {
	obj, err := k.client.Query(ctx, id)
	if err != nil {
		return err
	}
	return obj.Delete(ctx)
}

// ICS 03

func (k Keeper) OpenConnection(ctx sdk.Context,
	id, counterpartyID, clientID, counterpartyClientID string,
) error {
	obj, err := k.connection.Create(ctx, id, connection.Connection{
		Counterparty:       counterpartyID,
		Client:             clientID,
		CounterpartyClient: counterpartyClientID,
	})
	if err != nil {
		return err
	}
	return obj.Open(ctx)
}

/*
func (k Keeper) OpenTryConnection(ctx sdk.Context,
	id, counterpartyID, clientID, counterpartyClientID string,
	timeoutHeight, nextTimeoutHeight uint64,
	proofs []commitment.Proof,
) error {
	obj, err := k.connection.Create(ctx, id, connection.Connection{
		Counterparty:       counterpartyID,
		Client:             clientID,
		CounterpartyClient: counterpartyClientID,
	})
	if err != nil {
		return err
	}
	return k.ProofExec(ctx, id, proofs, func(ctx sdk.Context) error {
		return obj.OpenTry(ctx, timeoutHeight, nextTimeoutHeight)
	})
}

func (k Keeper) OpenAckConnection(ctx sdk.Context,
	id string,
	timeoutHeight, nextTimeoutHeight uint64,
	proofs []commitment.Proof,
) error {
	obj, err := k.connection.Query(ctx, id)
	if err != nil {
		return err
	}
	return k.ProofExec(ctx, id, proofs, func(ctx sdk.Context) error {
		return obj.OpenAck(ctx, timeoutHeight, nextTimeoutHeight)
	})
}

func (k Keeper) OpenConfirmConnection(ctx sdk.Context,
	id string,
	timeoutHeight uint64,
	proofs []commitment.Proof,
) error {
	obj, err := k.connection.Query(ctx, id)
	if err != nil {
		return err
	}
	return k.ProofExec(ctx, id, proofs, func(ctx sdk.Context) error {
		return obj.OpenConfirm(ctx, timeoutHeight)
	})
}
*/

func (k Keeper) QueryConnection(ctx sdk.Context, id string) (connection.Connection, error) {
	obj, err := k.connection.Query(ctx, id)
	if err != nil {
		return connection.Connection{}, err
	}
	return obj.Value(ctx), nil
}

// ICS 4

func (k Keeper) OpenChannel(ctx sdk.Context,
	moduleID, connID, id, counterpartyID, counterpartyModuleID string,
) error {
	obj, err := k.channel.Create(ctx, connID, id, channel.Channel{
		Module:             moduleID,
		Counterparty:       counterpartyID,
		CounterpartyModule: counterpartyModuleID,
	})
	if err != nil {
		return err
	}

	return obj.Open(ctx)
}

func (k Keeper) QueryChannel(ctx sdk.Context,
	connID, id string,
) (channel channel.Channel, err error) {
	obj, err := k.channel.Query(ctx, connID, id)
	if err != nil {
		return
	}
	return obj.Value(ctx), err
}

func (k Keeper) Send(ctx sdk.Context, connID, id string, packet channel.Packet) error {
	obj, err := k.channel.Query(ctx, connID, id)
	if err != nil {
		return err
	}
	return obj.Send(ctx, packet)
}

func (k Keeper) Receive(ctx sdk.Context, connID, id string, packet channel.Packet, proofs []commitment.Proof) error {
	obj, err := k.channel.Query(ctx, connID, id)
	if err != nil {
		return err
	}
	return k.ProofExec(ctx, connID, proofs, func(ctx sdk.Context) error {
		return obj.Receive(ctx, packet)
	})
}
